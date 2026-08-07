#!/usr/bin/env python3
"""Convert one DOCX or PDF file to Markdown and emit one JSON result.

Output layout:

    <source-dir>/<source-stem>.md
    <source-dir>/assets/<source-stem>.md/<image-file>

Standard output contains exactly one JSON object. The default process exit code
is zero even when conversion fails so llm_runner returns stdout unchanged.
Callers must inspect the JSON "success" field. Use --strict-exit-code for
conventional local CLI failure semantics.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import shutil
import statistics
import subprocess
import sys
import tempfile
from pathlib import Path


if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8")


SUPPORTED = {".docx", ".pdf"}
PDF_MIN_TEXT_CHARS = 20
PDF_RENDER_DPI = 180


def safe_name(value: str) -> str:
    value = re.sub(r"[\\/:*?\"<>|\x00-\x1f]", "-", value).strip(" .-")
    return value or "document"


def unique_asset_path(asset_dir: Path, stem: str, extension: str, payload: bytes) -> Path:
    digest = hashlib.sha256(payload).hexdigest()[:12]
    extension = re.sub(r"[^a-zA-Z0-9]", "", extension.lower()) or "bin"
    return asset_dir / f"{safe_name(stem)}-{digest}.{extension}"


def markdown_relative_path(markdown_path: Path, target: Path) -> str:
    try:
        relative = os.path.relpath(target.resolve(), markdown_path.parent.resolve())
    except ValueError:
        return target.resolve().as_posix()
    return Path(relative).as_posix()


def docx_heading_level(paragraph) -> int | None:
    style = (getattr(getattr(paragraph, "style", None), "name", "") or "").lower().replace(" ", "")
    match = re.match(r"heading([1-6])$", style)
    if match:
        return int(match.group(1))
    p_pr = paragraph._element.pPr
    if p_pr is not None and p_pr.outlineLvl is not None:
        return min(int(p_pr.outlineLvl.val) + 1, 6)
    return None


def docx_paragraph_markdown(paragraph, asset_dir: Path, markdown_path: Path, image_no: list[int]) -> str:
    from docx.oxml.ns import qn

    pieces: list[str] = []
    contents = paragraph.iter_inner_content() if hasattr(paragraph, "iter_inner_content") else paragraph.runs
    for run in contents:
        if hasattr(run, "url"):
            text = (run.text or "").strip()
            url = (run.url or "").strip()
            if text:
                pieces.append(f"[{text}]({url})" if url else text)
            continue
        text = run.text or ""
        if text:
            if run.bold and run.italic:
                text = f"***{text}***"
            elif run.bold:
                text = f"**{text}**"
            elif run.italic:
                text = f"*{text}*"
            pieces.append(text)
        for node in run._element.iter():
            if node.tag != qn("a:blip"):
                continue
            rel_id = node.get(qn("r:embed"))
            if not rel_id or rel_id not in paragraph.part.related_parts:
                continue
            part = paragraph.part.related_parts[rel_id]
            payload = part.blob
            ext = Path(str(part.partname)).suffix.lstrip(".") or "png"
            image_no[0] += 1
            target = unique_asset_path(asset_dir, f"image-{image_no[0]:04d}", ext, payload)
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(payload)
            relative = markdown_relative_path(markdown_path, target)
            pieces.append(f"![{paragraph.text.strip() or target.stem}]({relative})")
    text = "".join(pieces).strip()
    if not text:
        return ""
    level = docx_heading_level(paragraph)
    if level:
        return f"{'#' * level} {text}"
    p_pr = paragraph._element.pPr
    if p_pr is not None and p_pr.numPr is not None:
        indent = "  " * int(p_pr.numPr.ilvl.val if p_pr.numPr.ilvl is not None else 0)
        return f"{indent}- {text}"
    return text


def docx_table_markdown(table, asset_dir: Path, markdown_path: Path, image_no: list[int]) -> str:
    rows: list[list[str]] = []
    for row in table.rows:
        values: list[str] = []
        for cell in row.cells:
            paragraphs = [
                docx_paragraph_markdown(item, asset_dir, markdown_path, image_no)
                for item in cell.paragraphs
            ]
            values.append("<br>".join(value.replace("\n", "<br>") for value in paragraphs if value))
        rows.append(values)
    if not rows:
        return ""
    width = max(len(row) for row in rows)
    rows = [row + [""] * (width - len(row)) for row in rows]
    output = ["| " + " | ".join(rows[0]) + " |", "| " + " | ".join(["---"] * width) + " |"]
    output.extend("| " + " | ".join(row) + " |" for row in rows[1:])
    return "\n".join(output)


def convert_docx(path: Path, markdown_path: Path, asset_dir: Path) -> str:
    try:
        from docx import Document
        from docx.table import Table
        from docx.text.paragraph import Paragraph
    except ModuleNotFoundError as exc:
        raise RuntimeError("python-docx is required for DOCX conversion") from exc

    document = Document(path)
    blocks: list[str] = []
    image_no = [0]
    for child in document.element.body.iterchildren():
        if child.tag.endswith("}p"):
            value = docx_paragraph_markdown(Paragraph(child, document), asset_dir, markdown_path, image_no)
        elif child.tag.endswith("}tbl"):
            value = docx_table_markdown(Table(child, document), asset_dir, markdown_path, image_no)
        else:
            continue
        if value:
            blocks.append(value)
    return "\n\n".join(blocks).strip() + "\n"


def compact_pdf_text(value: str) -> str:
    return re.sub(r"[ \t]+", " ", value.replace("\r", "")).strip()


def pdf_heading_level(text: str, font_size: float, body_size: float) -> int | None:
    if not text or len(text) > 120:
        return None
    numeric = re.match(r"^(\d+(?:\.\d+)*)\s*([、.．])\s*\S+", text)
    if numeric:
        base_level = numeric.group(1).count(".") + 1
        if font_size >= body_size * 1.35 or (font_size <= 0 and numeric.group(2) == "、"):
            return min(base_level, 6)
        if font_size > 0 and numeric.group(2) == "、":
            return min(base_level + 1, 6)
        return None
    if font_size <= 0:
        return None
    if font_size >= body_size * 1.6:
        return 1
    if font_size >= body_size * 1.35:
        return 2
    if font_size >= body_size * 1.15:
        return 3
    return None


def pdf_text_markdown(page) -> tuple[str, int]:
    fragments: list[tuple[str, float]] = []

    def visitor(text, _cm, tm, _font, font_size):
        value = compact_pdf_text(str(text))
        if value:
            fragments.append((value, float(font_size or 0)))

    page.extract_text(visitor_text=visitor)
    fallback = page.extract_text(extraction_mode="layout") or ""
    text_count = len(re.sub(r"\s+", "", fallback))
    if text_count == 0:
        return "", 0
    sizes = [item[1] for item in fragments if item[1] > 0]
    body_size = statistics.median(sizes) if sizes else 10.0
    output: list[str] = []
    previous_blank = True
    for raw_line in fallback.splitlines():
        text = compact_pdf_text(raw_line)
        if not text:
            if output and not previous_blank:
                output.append("")
            previous_blank = True
            continue
        normalized = compact_pdf_text(text)
        numeric_line = bool(re.match(r"^\d+(?:\.\d+)*\s*[、.．]\s*\S+", normalized))
        matching_sizes = [
            size
            for fragment, size in fragments
            if fragment == normalized or (numeric_line and len(fragment) >= 4 and fragment in normalized)
        ]
        level = pdf_heading_level(text, max(matching_sizes, default=0), body_size)
        if level:
            text = f"{'#' * level} {text}"
        output.append(text)
        previous_blank = False
    value = "\n".join(output).strip()
    return value, text_count


def render_pdf_page(path: Path, page_number: int, asset_dir: Path) -> Path:
    executable = os.environ.get("PDFTOPPM_BIN") or shutil.which("pdftoppm")
    if not executable:
        raise RuntimeError("pdftoppm is required to render scanned PDF pages")
    asset_dir.mkdir(parents=True, exist_ok=True)
    prefix = asset_dir / f".__render-page-{page_number:04d}"
    rendered = Path(str(prefix) + ".png")
    rendered.unlink(missing_ok=True)
    command = [
        executable,
        "-f",
        str(page_number),
        "-l",
        str(page_number),
        "-singlefile",
        "-r",
        str(PDF_RENDER_DPI),
        "-png",
        str(path),
        str(prefix),
    ]
    result = subprocess.run(command, capture_output=True, text=True, timeout=120, check=False)
    if result.returncode != 0 or not rendered.is_file():
        rendered.unlink(missing_ok=True)
        detail = (result.stderr or result.stdout or "unknown pdftoppm error").strip()
        raise RuntimeError(f"cannot render PDF page {page_number}: {detail}")
    payload = rendered.read_bytes()
    rendered.unlink(missing_ok=True)
    target = unique_asset_path(asset_dir, f"page-{page_number}-scan", "png", payload)
    target.write_bytes(payload)
    return target


def pdf_page_render_markdown(path: Path, page_number: int, asset_dir: Path, markdown_path: Path) -> str:
    target = render_pdf_page(path, page_number, asset_dir)
    relative = markdown_relative_path(markdown_path, target)
    return f"![PDF page {page_number} render]({relative})"


def extract_pdf_images(page) -> list[tuple[str, bytes]]:
    images: list[tuple[str, bytes]] = []
    for image in page.images:
        payload = image.data
        extension = Path(image.name).suffix.lstrip(".") or "bin"
        if payload:
            images.append((extension, payload))
    return images


def convert_pdf(path: Path, markdown_path: Path, asset_dir: Path) -> str:
    try:
        from pypdf import PdfReader
    except ModuleNotFoundError as exc:
        raise RuntimeError("pypdf is required for PDF conversion") from exc

    document = PdfReader(str(path))
    output: list[str] = []
    image_no = 0
    for page_index, page in enumerate(document.pages):
        page_number = page_index + 1
        page_blocks: list[str] = [f"<!-- source-page: {page_number} -->"]
        page_text, text_count = pdf_text_markdown(page)
        if text_count == 0:
            page_blocks.extend(
                [
                    f"# PDF page {page_number} (image)",
                    "<!-- source-page-mode: image -->",
                    pdf_page_render_markdown(path, page_number, asset_dir, markdown_path),
                ]
            )
            output.append("\n\n".join(page_blocks))
            continue
        if page_text:
            page_blocks.append(page_text)
        if text_count < PDF_MIN_TEXT_CHARS:
            page_blocks.append(pdf_page_render_markdown(path, page_number, asset_dir, markdown_path))
            output.append("\n\n".join(page_blocks))
            continue
        try:
            images = extract_pdf_images(page)
        except Exception as exc:
            print(f"WARN: cannot enumerate images on page {page_number}: {exc}", file=sys.stderr, flush=True)
            images = []
        for extension, payload in images:
            image_no += 1
            target = unique_asset_path(asset_dir, f"page-{page_number}-image-{image_no:04d}", extension, payload)
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(payload)
            relative = markdown_relative_path(markdown_path, target)
            page_blocks.append(f"![Page {page_number} image {image_no}]({relative})")
        output.append("\n\n".join(page_blocks))
    if not output:
        raise ValueError(f"PDF contains no pages: {path.name}")
    return "\n\n".join(output).strip() + "\n"


def absolute_posix(path: Path) -> str:
    return path.resolve().as_posix()


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def file_record(path: Path, kind: str, base_dir: Path) -> dict[str, object]:
    return {
        "type": kind,
        "name": path.name,
        "extension": path.suffix.lower(),
        "path": absolute_posix(path),
        "relative_path": Path(os.path.relpath(path, base_dir)).as_posix(),
        "size": path.stat().st_size,
        "sha256": sha256_file(path),
    }


def publish_outputs(
    temporary_root: Path,
    temporary_markdown: Path,
    temporary_assets: Path,
    markdown_path: Path,
    asset_dir: Path,
    has_assets: bool,
) -> None:
    assets_root = asset_dir.parent
    assets_root.mkdir(parents=True, exist_ok=True)
    previous_assets = temporary_root / "previous-assets"

    if asset_dir.exists():
        asset_dir.replace(previous_assets)

    try:
        if has_assets:
            temporary_assets.replace(asset_dir)
        else:
            shutil.rmtree(temporary_assets)
        os.replace(temporary_markdown, markdown_path)
    except Exception:
        if asset_dir.is_dir():
            shutil.rmtree(asset_dir)
        elif asset_dir.exists():
            asset_dir.unlink()
        if previous_assets.exists():
            previous_assets.replace(asset_dir)
        raise

    if not has_assets:
        try:
            assets_root.rmdir()
        except OSError:
            pass


def convert_single(path: Path) -> dict[str, object]:
    source_path = path.expanduser().resolve()
    if not source_path.is_file():
        raise ValueError(f"input file does not exist: {source_path}")
    if source_path.suffix.lower() not in SUPPORTED:
        raise ValueError(f"unsupported file type: {source_path.suffix or '<none>'}; expected DOCX or PDF")

    markdown_path = source_path.with_suffix(".md")
    asset_dir = source_path.parent / "assets" / markdown_path.name
    temporary_root = Path(tempfile.mkdtemp(prefix=".document-convert-", dir=source_path.parent))

    try:
        temporary_markdown = temporary_root / markdown_path.name
        temporary_assets = temporary_root / "assets" / markdown_path.name
        temporary_assets.mkdir(parents=True, exist_ok=True)

        if source_path.suffix.lower() == ".docx":
            content = convert_docx(source_path, temporary_markdown, temporary_assets)
        else:
            content = convert_pdf(source_path, temporary_markdown, temporary_assets)

        temporary_markdown.write_text(content, encoding="utf-8", newline="\n")
        temporary_asset_files = sorted(item for item in temporary_assets.rglob("*") if item.is_file())
        publish_outputs(
            temporary_root,
            temporary_markdown,
            temporary_assets,
            markdown_path,
            asset_dir,
            bool(temporary_asset_files),
        )

        image_paths = sorted(item for item in asset_dir.rglob("*") if item.is_file()) if asset_dir.is_dir() else []
        markdown_record = file_record(markdown_path, "markdown", source_path.parent)
        image_records = [file_record(item, "image", source_path.parent) for item in image_paths]
        copy_paths = [markdown_record["path"], *(record["path"] for record in image_records)]

        return {
            "success": True,
            "source_path": absolute_posix(source_path),
            "markdown_path": markdown_record["path"],
            "assets_dir": absolute_posix(asset_dir),
            "image_paths": [record["path"] for record in image_records],
            "copy_paths": copy_paths,
            "file_count": len(copy_paths),
            "markdown_char_count": len(content),
            "files": [markdown_record, *image_records],
        }
    finally:
        shutil.rmtree(temporary_root, ignore_errors=True)


def error_result(input_value: str, exc: Exception) -> dict[str, object]:
    return {
        "success": False,
        "source_path": input_value,
        "markdown_path": "",
        "assets_dir": "",
        "image_paths": [],
        "copy_paths": [],
        "file_count": 0,
        "files": [],
        "error": {
            "type": type(exc).__name__,
            "message": str(exc),
        },
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, help="one DOCX or PDF file")
    parser.add_argument(
        "--strict-exit-code",
        action="store_true",
        help="return exit code 1 on failure; disabled by default for llm_runner JSON responses",
    )
    args = parser.parse_args()

    input_value = args.input.as_posix() if args.input is not None else ""
    try:
        if args.input is None:
            raise ValueError("--input is required")
        result = convert_single(args.input)
        exit_code = 0
    except Exception as exc:
        result = error_result(input_value, exc)
        exit_code = 1 if args.strict_exit_code else 0

    print(json.dumps(result, ensure_ascii=False, separators=(",", ":")), flush=True)
    return exit_code


if __name__ == "__main__":
    raise SystemExit(main())
