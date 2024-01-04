#!/bin/bash

# Navigate to the project's source directory
cd ./app/

# Install Babel, Less, and LightningCSS for JS and CSS processing
yes | npm install --silent @babel/core @babel/cli @babel/preset-env
yes | npm install --silent less lightningcss-cli

# Compile LESS files into one unique CSS file
npx --yes lessc ./client/assets/css/style.less > ./client/assets/css/tmp.css

# Minify and Prefix CSS
npx --yes lightningcss --minify --bundle --targets 'cover 99.5%' ./client/assets/css/tmp.css -o ./client/assets/css/style.css

# Save the original JS file
cp ./client/assets/js/isaiah.js ./client/assets/js/isaiah.backup.js

# Make JS cross-browser-compatible
npx --yes babel ./client/assets/js/isaiah.js --out-file ./client/assets/js/isaiah.js --config-file ./.babelrc.json

# Minify JS
npx --yes terser ./client/assets/js/isaiah.js -o ./client/assets/js/isaiah.js

# Build the app
go build -o isaiah main.go

# Reset CSS and JS
rm -f ./client/assets/css/tmp.css
rm -f ./client/assets/css/style.css
mv ./client/assets/js/isaiah.backup.js ./client/assets/js/isaiah.js

# Remove any previous installation
rm -f /usr/bin/isaiah

# Install the app's binary 
mv isaiah /usr/bin/
chmod 755 /usr/bin/isaiah
