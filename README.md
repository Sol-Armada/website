# Sol Armada Website

## Build logo

Build logo as png using `inkscape`

```
inkscape -w 300 -h 300 --export-id logo-white -o ./src/assets/logo.png ./src/assets/logo.svg
inkscape -w 300 -h 300 --export-id logo-white -o ./src/assets/logo-white.png ./src/assets/logo.svg
inkscape -w 300 -h 300 --export-id logo-blue -o ./src/assets/logo-blue.png ./src/assets/logo.svg
inkscape -w 300 -h 300 --export-id logo-dark-grey -o ./src/assets/logo-dark-grey.png ./src/assets/logo.svg


```

Conver the logo to a favicon using `inkscape` and `imagemagick`

```
inkscape -w 16 -h 16 --export-id logo-white -o ./src/assets/logo-16.png ./src/assets/logo.svg
inkscape -w 32 -h 32 --export-id logo-white -o ./src/assets/logo-32.png ./src/assets/logo.svg
inkscape -w 48 -h 48 --export-id logo-white -o ./src/assets/logo-48.png ./src/assets/logo.svg

convert ./src/assets/logo-16.png ./src/assets/logo-32.png ./src/assets/logo-48.png ./public/favicon.ico
```
