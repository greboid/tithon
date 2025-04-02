const { exec } = require('child_process');
const fs = require('fs');
const path = require('path')
module.exports = {
  rebuildConfig: {
    force: true
  },
  hooks: {
    generateAssets: async (forgeConfig, platform, arch) => {
      const child = exec('go build -o backend '+__dirname, (err) => {
        if (err) {
          console.log("Error building backend")
          throw err
        }
      })
      await new Promise((resolve) => { child.on('close', resolve)})
    },
    packageAfterExtract: async (forgeConfig, buildPath, electronVersion, platform, arch) => {
      fs.copyFile(path.join(__dirname, 'backend'), path.join(buildPath, 'backend'), (err) => {
        if (err) {
          console.log("Error moving backend")
          throw err
        }
      })
    }
  },
  packagerConfig: {
    name: 'goircha',
    icon: 'web/static/icon.png'
  },
  makers: [
    {
      name: '@electron-forge/maker-zip'
    },
    {
      name: '@electron-forge/maker-deb',
      platforms: ['linux'],
    }
  ]
};
