const {app, BrowserWindow, shell} = require('electron')
const { spawn } = require('child_process');
const {join} = require('node:path')

const createWindow = async () => {
  const win = new BrowserWindow({
                                  icon: 'web/static/icon.png',
                                  width:  800,
                                  height: 600,
                                })
  const child = spawn(join(__dirname, 'backend'))
  child.on('exit', () => {
    app.quit()
  })
  child.stdout.once('data', () => {
    win.loadURL('http://localhost:8081')
       .catch(() => app.quit())
  })
  win.setMenu(null)
  win.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: 'deny' };
  });
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
