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

# Append a version parameter to the main JS & CSS linked files to prevent caching
VERSION=$(git describe --tags --abbrev=0)
sed -i.bak "s/isaiah.js/isaiah.js?v=$VERSION/" ./client/index.html
sed -i.bak "s/style.css/style.css?v=$VERSION/" ./client/index.html
sed -i.bak "s/-VERSION-/$VERSION/" ./client/assets/js/isaiah.js

# Replace the version tag with the current version in the main Go file
sed -i.bak "s/-VERSION-/$VERSION/" ./main.go

# Build the app
go build -o isaiah main.go

# Reset CSS and JS
rm -f ./client/assets/css/tmp.css
rm -f ./client/assets/css/style.css
mv ./client/assets/js/isaiah.backup.js ./client/assets/js/isaiah.js

# Remove backup files
rm -f ./client/index.html.bak
rm -f ./client/assets/js/isaiah.js.bak

DESTINATION="/usr/bin"
if [ -d "/usr/local/bin" ]; then
  DESTINATION="/usr/local/bin"
fi

# Remove any previous installation
rm -f $DESTINATION/isaiah

# Install the app's binary 
mv isaiah $DESTINATION/
chmod 755 $DESTINATION/isaiah
