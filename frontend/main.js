const {app, globalShortcut, BrowserWindow, Menu, MenuItem, shell} = require('electron')
const {spawn} = require('child_process')
const {join} = require('node:path')

let child
let globalPort

const parsePort = (win) => {
  return async (data) => {
    const text = new TextDecoder().decode(data)
    const { port } = /port=(?<port>\d+)/.exec(text).groups
    globalPort = port
    win.loadURL(`http://localhost:${globalPort}`)
       .catch(quit)
  }
}

const outputLogs = async (data) => {
  const text = new TextDecoder().decode(data)
  console.log(text)
}

const quit = (error) => {
  console.log(`Quitting: ${error}`)
  app.quit()
}

const createWindow = async () => {
  let port = -1;
  const win = new BrowserWindow({
                                  icon:           'icon.png',
                                  width:          800,
                                  height:         600,
                                  webPreferences: {
                                    autoplayPolicy:       'no-user-gesture-required',
                                    backgroundThrottling: false,
                                    defaultEncoding:      'UTF-8',
                                  },
                                })
  child = spawn(join(__dirname, 'backend'), [], {windowsHide: false})
  child.on('exit', quit)
  child.stdout.once('data', parsePort(win))
  child.stdout.on('data', outputLogs)
  win.webContents.setWindowOpenHandler(({url}) => {
    shell.openExternal(url)
    return {action: 'deny'}
  })
  const menu = new Menu()
  menu.append(new MenuItem({
                             label:       'Refresh',
                             accelerator: 'F5',
                             click:       () => {
                               win.loadURL(`http://localhost:${globalPort}`)
                                  .catch(e => {
                                    console.log(e)
                                    app.quit()
                                  })
                             },
                           }))
  menu.append(new MenuItem({
                             label:       'Show Dev Tools',
                             accelerator: 'F12',
                             click:       () => {
                               win.webContents.toggleDevTools()
                             },
                           }))
  Menu.setApplicationMenu(menu)
  win.setMenuBarVisibility(false)
}
app.setName('tithon')
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
