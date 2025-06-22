const handlePaste = async (e) => {
  for (const clipboardItem of e.clipboardData.files) {
    if (clipboardItem.type.startsWith('image/')) {
      console.log(clipboardItem)
    }
  }
}

window.addEventListener('paste', handlePaste)

const notify = (title, text, popup, sound, noise) => {
  if (popup) {
    new Notification(title, {
      body: text,
      icon: "/static/icon.png"
    });
  }
  if (sound) {
    if (!noise) {
      noise = "/static/notification.mp3"
    }
    new Audio(noise).play()
  }
}

Notification.requestPermission().catch(e => console.log(e))
