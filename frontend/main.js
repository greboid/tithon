const {app, globalShortcut, BrowserWindow, Menu, MenuItem, shell} = require('electron')
const {spawn} = require('child_process')
const {join} = require('node:path')

let child
let globalPort

const parsePort = (win) => {
  return async (data) => {
    const text = new TextDecoder().decode(data)
    const {port} = /port=(?<port>\d+)/.exec(text).groups
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
  const win = new BrowserWindow(
      {
        icon:           'icon.png',
        width:          800,
        height:         600,
        webPreferences: {
          autoplayPolicy:       'no-user-gesture-required',
          backgroundThrottling: false,
          defaultEncoding:      'UTF-8',
          spellcheck:           true,
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
  menu.append(new MenuItem(
      {
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
  menu.append(new MenuItem(
      {
        label:       'Show Dev Tools',
        accelerator: 'F12',
        click:       () => {
          win.webContents.toggleDevTools()
        },
      }))
  Menu.setApplicationMenu(menu)
  win.setMenuBarVisibility(false)

  win.webContents.on('context-menu', (event, params) => {
    if (params.isEditable || params.editFlags.canSelectAll) {
      const menu = Menu.buildFromTemplate(
          [
            {role: 'cut', visible: params.editFlags.canCut},
            {role: 'copy', visible: params.editFlags.canCopy},
            {role: 'paste', visible: params.editFlags.canPaste},
            {role: 'delete', visible: params.editFlags.canDelete},
            {role: 'selectall', visible: params.editFlags.canCopy},
          ])
      if (params.dictionarySuggestions.length > 0 || params.misspelledWord) {
        menu.append(new MenuItem({type: 'separator'}))
        for (const suggestion of params.dictionarySuggestions) {
          menu.append(new MenuItem(
              {
                label: suggestion,
                click: () => win.webContents.replaceMisspelling(suggestion),
              }))
        }
        if (params.misspelledWord) {
          menu.append(
              new MenuItem(
                  {
                    label: 'Add to dictionary',
                    click: () => win.webContents.session.addWordToSpellCheckerDictionary(params.misspelledWord),
                  }),
          )
        }
      }
      if (menu.items.some((item) => item.visible)) {
        menu.popup()
      }
    }
  })
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
