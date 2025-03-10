let atBottom = false
window.addEventListener('scroll', function(event) {
  if (event.target.id === 'messages') {
    atBottom = event.target.scrollTop === event.target.scrollTopMax
  }
}, true)
const typingCallback = (e) => {
  if (e.key === "Tab") {
    e.preventDefault()
  }
  if (e.key.length !== 1) {
    return
  }
  if (e.key === "c") {
    return
  }
  document.getElementById("textInput").focus()
}
const config = {childList: true, subtree: true}
const textCallback = mutations => {
  for (const mutation of mutations) {
    if (mutation.addedNodes.length > 0) {
      for (const node of mutation.addedNodes) {
        if (atBottom) {
          node.scrollIntoView({ behavior: "smooth", block: "end", inline: "nearest" })
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
        }
      }
    }
  }
}
let observer = new MutationObserver(globalCallback)
observer.observe(document, config)
window.addEventListener('keydown', typingCallback)
