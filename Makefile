buildwindows:
	fyne package -os windows -icon resources\goCryptor.png

build:
	fyne package -os linux -icon resources\goCryptor.png

buildmobile:
	fyne package -os android -appID com.example.myapp -icon resources\goCryptor.png
	fyne package -os ios - appID com.example.myapp -icon resources\goCryptor.png