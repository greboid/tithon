let atBottom = true
window.addEventListener('scroll', function(event) {
  if (event.target.id === 'messages') {
    const target = event.target
    atBottom = (target.scrollHeight - target.scrollTop - target.clientHeight) < 5
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

const handlePaste = async (e) => {
  for (const clipboardItem of e.clipboardData.files) {
    if (clipboardItem.type.startsWith('image/')) {
      console.log(clipboardItem)
    }
  }
}

window.addEventListener('paste', handlePaste)

const notify = (text, noise) => {
  new Notification("Tithon", { body: text, icon: "/static/icon.png", data: {
    channel: "#mdbot"
    }});
  if (!noise) {
    noise = "/static/notification.mp3"
  }
  new Audio(noise).play()
}

Notification.requestPermission().catch(e => console.log(e))
