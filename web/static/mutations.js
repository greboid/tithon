let atBottom = false
window.addEventListener('scroll', function(event) {
  if (event.target.id === 'window') {
    atBottom = event.target.scrollTop === event.target.scrollTopMax
  }
}, true)
const config = {childList: true, subtree: true}
const textCallback = mutations => {
  for (const mutation of mutations) {
    if (mutation.addedNodes.length > 0) {
      for (const node of mutation.addedNodes) {
        if (atBottom) {
          node.scrollIntoView()
        }
      }
    }
  }
}
const globalCallback = mutations => {
  for (const mutation of mutations) {
    if (mutation.addedNodes.length > 0) {
      for (const node of mutation.addedNodes) {
        if (node.id === 'text') {
          observer.disconnect()
          observer = new MutationObserver(textCallback)
          observer.observe(node, config)
        }
      }
    }
  }
}
let observer = new MutationObserver(globalCallback)
observer.observe(document, config)
