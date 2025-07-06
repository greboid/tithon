const handlePaste = async (e) => {
  const fileUpload = document.getElementById('fileUpload')
  if (!fileUpload) return
  
  for (const clipboardItem of e.clipboardData.files) {
      const list = new DataTransfer()
      list.items.add(clipboardItem)
      fileUpload.files = list.files
      const changeEvent = new Event('change', { bubbles: true })
      fileUpload.dispatchEvent(changeEvent)
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
