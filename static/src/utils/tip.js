// 轻量提示，替代 element-plus 的 ElMessage（避免引入整套 UI 库）
let container = null

function ensureContainer() {
  if (container) return container
  container = document.createElement('div')
  container.className = 'toast-container'
  document.body.appendChild(container)
  return container
}

function showToast(message, type) {
  const c = ensureContainer()
  const el = document.createElement('div')
  el.className = 'toast toast-' + (type || 'info')
  el.textContent = message
  c.appendChild(el)
  // 触发入场动画
  requestAnimationFrame(() => el.classList.add('show'))
  setTimeout(() => {
    el.classList.remove('show')
    setTimeout(() => el.remove(), 300)
  }, 2400)
}

// 使用防抖函数包装，1s 内只能发送一次消息
export function throttle(fn, interval) {
  let last = 0
  return function (...args) {
    const now = Date.now()
    if (now - last >= interval) {
      last = now
      fn.apply(this, args)
    }
  }
}

export const sendErrorMessage = throttle((message) => showToast(message, 'error'), 1000)
export const sendSuccessMessage = throttle((message) => showToast(message, 'success'), 1000)
