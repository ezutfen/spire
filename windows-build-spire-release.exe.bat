START /wait taskkill /f /im spire.exe

:: Install node env
call npm install -g win-node-env

:: Build SPA (Frontend)
cd frontend && call npm run build & cd ..

:: Copy local asset images
:: curl -o %localappdata%\Temp\assets.zip -L https://github.com/Akkadius/eq-asset-preview/archive/refs/heads/master.zip
:: unzip -o %localappdata%\Temp\assets.zip -d %localappdata%\Temp\assets
:: xcopy "%localappdata%\Temp\assets\eq-asset-preview-master\assets\" "frontend\dist\assets\" /s /e /y

@REM xcopy "frontend\dist\" "public\" /s /e /y

go build -o spire.exe
