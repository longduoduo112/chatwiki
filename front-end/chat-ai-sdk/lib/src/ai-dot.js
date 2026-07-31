import AiAvatar from "./ai-avatar";

class AiDot {
  dotEl = null;

  create(config) {
    if(config.value <= 0){
      this.remove()
      return;
    }

    if(config.value > 0 && !this.dotEl){
      if (!AiAvatar.avatarContentEl) {
        return
      }
      this.dotEl = document.createElement("div");
      AiAvatar.avatarContentEl.appendChild(this.dotEl);
    }

    if(this.dotEl){
      this.dotEl.className = `ai-dot${Number(config.value) > 9 ? ' ai-dot-plus' : ''}`;
      this.dotEl.textContent = config.value;
    }
  }

  remove() {
    if (this.dotEl) {
      this.dotEl.remove();
      this.dotEl = null;
    }
  }
}

export default new AiDot();
