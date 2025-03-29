const {app, BrowserWindow} = require('electron')
const { spawn } = require('child_process');
const {join} = require('node:path')

const sleep = ms => new Promise(resolve => setTimeout(resolve, ms))

const createWindow = async () => {
  const win = new BrowserWindow({
                                  width:  800,
                                  height: 600,
                                })
  const child = spawn(join(__dirname, 'ircclient'))
  child.on('exit', () => {
    console.log("Exited")
  })
  child.stdout.once('data', () => {
    win.loadURL('http://localhost:8081')
       .catch(err => console.log(err))
  })
  win.setMenu(null)
}
app.commandLine.appendSwitch('disable-http-cache')
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})
app.whenReady()
   .then(async () => {
     await createWindow()
     app.on('activate', () => {
       if (BrowserWindow.getAllWindows().length === 0) {createWindow()}
     })
   })
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {app.quit()}
})
