const {app, globalShortcut, BrowserWindow, shell} = require('electron')
const { register } = require("electron-shortcuts")
const {spawn} = require('child_process')
const {join} = require('node:path')

let child

const createWindow = async () => {
  const win = new BrowserWindow({
                                  icon:   'icon.png',
                                  width:  800,
                                  height: 600,
                                })
  child = spawn(join(__dirname, 'backend'))
  child.on('exit', () => {
    app.quit()
  })
  child.stdout.once('data', () => {
    win.loadURL('http://localhost:8081')
       .catch(() => app.quit())
  })
  register('F5', () => {
    win.loadURL('http://localhost:8081')
       .catch(() => app.quit())
  }, win)
  register('F12', () => {
    win.webContents.openDevTools()
  }, win)
  // child.stdout.on('data', (data) => {
  //   console.log(new TextDecoder().decode(data))
  // })
  win.setMenu(null)
  win.webContents.setWindowOpenHandler(({url}) => {
    shell.openExternal(url)
    return {action: 'deny'}
  })
}
app.commandLine.appendSwitch('disable-http-cache')
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
    child.kill()
  }
})
app.whenReady()
   .then(async () => {
     await createWindow()
     app.on('activate', () => {
       if (BrowserWindow.getAllWindows().length === 0) {createWindow()}
     })
   })
