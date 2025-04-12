const { exec } = require('child_process');
const fs = require('fs');
const path = require('path')
module.exports = {
  rebuildConfig: {
    force: true
  },
  hooks: {
    generateAssets: async (config, platform, arch) => {
      const child = exec('go build -C '+path.join(__dirname, "..", "backend")+' -o '+path.join(__dirname, "backend")+' .', (err) => {
        if (err) {
          console.log(`Error building backend: ${err}`)
        }
      })
      await new Promise((resolve) => { child.on('close', resolve)})
    }
  },
  packagerConfig: {
    name: 'tithon',
    icon: 'icon.png',
    executable: "tithon",
  },
  makers: [
    {
      name: '@electron-forge/maker-zip'
    },
    {
      name: '@electron-forge/maker-deb',
      platforms: ['linux'],
      config: {
        options: {
          name: "tithon",
          productName: "Tithon IRC",
          maintainer: 'Greboid',
          homepage: 'https://github.com/greboid/tithon',
          icon: 'icon.png',
          section: "Network",
          categories: ['Network', 'Chat', 'IRCClient'],
          description: "Simple IRC Client",
          license: "MIT"
        }
      }
    }
  ],
  publishers: [
    {
      name: '@electron-forge/publisher-github',
      config: {
        repository: {
          owner: 'greboid',
          name: 'tithon'
        },
        prerelease: false,
        draft: false
      }
    }
  ]
};
