let atBottom = true
window.addEventListener('scroll', function(event) {
  if (event.target.id === 'messages') {
    atBottom = event.target.scrollTop === (event.target.scrollHeight-event.target.clientHeight)
  }
}, true)

const config = {childList: true, subtree: true}
const textCallback = mutations => {
  for (const mutation of mutations) {
    if (mutation.addedNodes.length > 0) {
      for (const node of mutation.addedNodes) {
        if (atBottom) {
          setTimeout(() => document.getElementById('messages').scrollTop = document.getElementById('messages').scrollHeight-document.getElementById('messages').clientHeight, 10)
        }
      }
    }
  }
}
const globalCallback = mutations => {
  for (const mutation of mutations) {
    if (mutation.addedNodes.length > 0) {
      for (const node of mutation.addedNodes) {
        if (node.id === 'messages') {
          observer.disconnect()
          observer = new MutationObserver(textCallback)
          observer.observe(node, config)
          setTimeout(() => document.getElementById('messages').scrollTop = document.getElementById('messages').scrollTopMax, 10)
        }
      }
    }
  }
}
let observer = new MutationObserver(globalCallback)
observer.observe(document, config)
