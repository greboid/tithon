const typingCallback = (e) => {
  if (e.key === "Tab") {
    if (e.target.id === "#textInput") {
      e.preventDefault()
    }
  }
  if (e.key.length !== 1) {
    return
  }
  if (e.key === "c") {
    return
  }
  document.getElementById("textInput").focus()
}
window.addEventListener('keydown', typingCallback)
