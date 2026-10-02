// naive-ui 的模态/卡片关闭按钮是内部渲染的图标按钮，没有暴露
// 设置 aria-label 的 props（X 按钮在读屏下是匿名控件）。上游未提供
// 配置点，这里用一个轻量观察器在它们进入 DOM 时补上可访问名称。
const CLOSE_PATTERN = /n-card-header__close|n-dialog__close|n-base-close/;

let installed = false;

function labelCloseButton(button: HTMLButtonElement): void {
  if (button.getAttribute('aria-label') || button.textContent?.trim()) return;
  if (CLOSE_PATTERN.test(button.className.toString())) {
    button.setAttribute('aria-label', '关闭');
  }
}

function labelCloseButtons(root: ParentNode): void {
  if (root instanceof HTMLButtonElement) {
    labelCloseButton(root);
    return;
  }
  const buttons = root.querySelectorAll?.('button');
  if (!buttons) return;
  for (const button of Array.from(buttons)) {
    labelCloseButton(button);
  }
}

export function installCloseButtonA11y(): void {
  if (installed || typeof MutationObserver === 'undefined') return;
  installed = true;
  labelCloseButtons(document.body);
  new MutationObserver((records) => {
    for (const record of records) {
      for (const node of Array.from(record.addedNodes)) {
        if (node instanceof HTMLElement) labelCloseButtons(node);
      }
    }
  }).observe(document.body, { childList: true, subtree: true });
}
