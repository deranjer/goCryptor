package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/deranjer/gocryptor/encryptor"
	"github.com/integrii/flaggy"
	"github.com/sqweek/dialog"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

type goCryptorUI struct {
	action           string
	statusLabel      *widget.Label
	passwordEntry    *widget.Entry
	passConfirmEntry *widget.Entry
	fileName         string
	fileNameLabel    *widget.Label
	overwriteFile    bool
}

func (ui *goCryptorUI) encryptFile() {
	err := ui.validateInformation()
	if err != nil {
		return
	}
	isDir, err := validateFileName(ui.fileName)
	if err != nil {
		ui.statusLabel.SetText("Error: " + err.Error())
		go ui.statusFade(3)
		return
	}
	if isDir {
		err = filepath.Walk(ui.fileName, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			err = encryptor.EncryptFile(ui.passwordEntry.Text, path)
			if err != nil {
				ui.statusLabel.SetText("Error encrypting file: " + err.Error())
				return nil
			}
			return nil
		})
		if err != nil {
			fmt.Println("Walk dir err: ", err)
		}
	} else {
		err = encryptor.EncryptFile(ui.passwordEntry.Text, ui.fileName)
		if err != nil {
			ui.statusLabel.SetText("Error encrypting file: " + err.Error())
			return
		}
	}
	ui.statusLabel.SetText("Success encrypting file(s)!")
	go ui.statusFade(5)
	ui.passwordEntry.SetText("")
	ui.passConfirmEntry.SetText("")
	ui.fileName = ""
	ui.fileNameLabel.SetText("File Path: ")
}

func (ui *goCryptorUI) decryptFile() {
	err := ui.validateInformation()
	if err != nil {
		return
	}
	err = encryptor.DecryptFile(ui.passwordEntry.Text, ui.fileName, ui.overwriteFile)
	if err != nil {
		ui.statusLabel.SetText("Error decrypting file: " + err.Error())
		go ui.statusFade(8)
		return
	}
	ui.statusLabel.SetText("Success decrypting file!")
	go ui.statusFade(8)
	ui.passwordEntry.SetText("")
	ui.passConfirmEntry.SetText("")
	ui.fileName = ""
	ui.fileNameLabel.SetText("File Path: ")
}

func (ui *goCryptorUI) validateInformation() error {
	errStatus := errors.New("information validation failed")
	if ui.passwordEntry.Text == "" {
		ui.statusLabel.SetText("Password cannot be empty!")
		go ui.statusFade(3)
		return errStatus
	}
	if ui.passwordEntry.Text == ui.passConfirmEntry.Text {
		ui.statusLabel.SetText("Confirmed passwords...")
		go ui.statusFade(3)
		return nil
	} else {
		ui.statusLabel.SetText("Passwords do not match! Please try again.")
		go ui.statusFade(3)
		return errStatus
	}
}

func (ui *goCryptorUI) statusFade(waitTime time.Duration) {
	time.Sleep(waitTime * time.Second)
	ui.statusLabel.SetText("")
}

func (ui *goCryptorUI) browseFile() string {
	filename, err := dialog.File().Title("Select File").Load()
	if err == dialog.ErrCancelled {
		return ""
	}
	if err != nil {
		ui.statusLabel.SetText("File picker failure: " + err.Error())
		ui.statusFade(4)
		os.Exit(0)
	}
	return filename
}

func (ui *goCryptorUI) browseFolder() string {
	folderName, err := dialog.Directory().Title("Select Directory").Browse()
	if err == dialog.ErrCancelled {
		return ""
	}
	if err != nil {
		ui.statusLabel.SetText("Folder picker failure: " + err.Error())
		ui.statusFade(4)
		os.Exit(0)
	}
	fmt.Println("FolderName: ", folderName)
	return folderName
}

func parseFlags() (string, string) {
	flaggy.SetName("goCryptor")
	flaggy.SetDescription("Encrypts and decrypts files and folders")
	flaggy.DefaultParser.ShowHelpOnUnexpected = true
	// set the default of encrypt
	var encryptFlag string
	// setup the encrypt flag
	flaggy.String(&encryptFlag, "e", "encrypt", "encrypt a file or folder")
	// decrypt var
	var decryptFlag string
	flaggy.String(&decryptFlag, "d", "decrypt", "selects file to decrypt")
	// parse the results
	flaggy.Parse()
	if encryptFlag != "" && decryptFlag != "" {
		fmt.Println("cannot perform both encrypt and decrypt in one run")
		os.Exit(0)
	}
	if encryptFlag != "" {
		fmt.Println("Encrypt file: ", encryptFlag)
		return "encrypt", encryptFlag
	}
	if decryptFlag != "" {
		return "decrypt", decryptFlag
	}
	return "", ""
}

// validateFileName checks a few things about the supplied name to make sure it is legit
func validateFileName(fileName string) (bool, error) {
	// if no filename at all supplied then just return to the main so the user can chose one
	if fileName == "" {
		return false, errors.New("no filename supplied")
	}
	// if a filename WAS supplied, make sure it is legit
	file, err := os.Stat(fileName)
	if err != nil {
		return false, err
	}
	if file.IsDir() {
		return true, nil
	}
	return false, nil
}

func main() {
	// Setup our main app struct
	ui := goCryptorUI{}
	// action attempts to automatically determine if we are encrypting or decrypting
	ui.action = "encrypt"
	// fileName is the name of the file or folder to encrypt
	actionText, fileName := parseFlags()
	ui.fileName = fileName
	if fileName != "" {
		_, err := validateFileName(fileName)
		if err != nil {
			fmt.Println("error reading file: ", err)
			os.Exit(0)
		}
	}
	mainApp := app.New()
	//mainApp.SetIcon()
	mainWindow := mainApp.NewWindow("goCryptor")
	mainWindow.SetFixedSize(true)
	mainWindow.CenterOnScreen()
	// Setup the text at the top of the app
	mainTitle := widget.NewLabelWithStyle("goCryptor", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	// Setup file selection area
	ui.fileNameLabel = widget.NewLabel("File Path: " + ui.fileName)
	// create the overwrite check box, and set it to true by default
	overwriteCheck := widget.NewCheck("Overwrite on Decrypt", func(bool) {
		if ui.overwriteFile == true {
			ui.overwriteFile = false
		} else {
			ui.overwriteFile = true
		}
	})
	overwriteCheck.SetChecked(true)
	ui.overwriteFile = true
	// File selection box
	selectBox := widget.NewHBox(
		widget.NewButton("Select File", func() {
			ui.fileName = ui.browseFile()
			ui.fileNameLabel.SetText("File Path: " + filepath.Base(ui.fileName))
		}),
		widget.NewButton("Select Folder", func() {
			ui.fileName = ui.browseFolder()
			ui.fileNameLabel.SetText("File Path: " + filepath.Base(ui.fileName))
		}),
		overwriteCheck,
	)

	// Setting up the form for the password entry
	ui.passwordEntry = widget.NewPasswordEntry()
	ui.passConfirmEntry = widget.NewPasswordEntry()
	// After setting up the input boxes, create the form
	passwordForm := widget.NewForm()
	// Use the append function to add in both of the inputs with labels
	passwordForm.Append("Password: ", ui.passwordEntry)
	passwordForm.Append("Confirm Password: ", ui.passConfirmEntry)
	// Setup the status message
	ui.statusLabel = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	// Setup scroll container for status message
	scrollSize := fyne.NewSize(400, 50)
	scrollContainer := widget.NewHScrollContainer(ui.statusLabel)
	scrollContainer.SetMinSize(scrollSize)
	// Setup Encrypt Button
	encryptButton := widget.NewButton("Encrypt", func() {
		ui.encryptFile()
	})
	encryptButton.Style = widget.PrimaryButton
	// Setup Decrypt Button
	decryptButton := widget.NewButton("Decrypt", func() {
		ui.decryptFile()
	})
	decryptButton.Style = widget.PrimaryButton
	// Using actionText to see if we are only performing a certain action (from command line)
	switch actionText {
	case "encrypt":
		decryptButton.Disable()
	case "decrypt":
		encryptButton.Disable()
	}
	// Create our URL
	url, err := url.Parse("https://github.com")
	if err != nil {
		panic(err)
	}
	// Put both of the Buttons in an Hbox with a spacer in between
	buttons := widget.NewHBox(
		widget.NewButton("Cancel", func() {
			mainWindow.Close()
		}),
		widget.NewHyperlink("About ", url),
		layout.NewSpacer(),
		decryptButton,
		encryptButton,
	)
	// Creating our full box
	fullBox := widget.NewVBox(
		mainTitle,
		layout.NewSpacer(),
		selectBox,
		ui.fileNameLabel,
		layout.NewSpacer(),
		passwordForm,
		scrollContainer,
		layout.NewSpacer(),
		buttons,
	)
	// Check for CAPS LOCK on the canvas (does not work on all elements)
	if deskCanvas, ok := mainWindow.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyUp(func(ev *fyne.KeyEvent) {
			if ev.Name == "CapsLock" {
				ui.statusLabel.SetText("CAPS LOCK key was toggled!")
				go ui.statusFade(2)
			}
		})
	}
	// Set our main layout and input our Vertical Box into it
	// Give the box a fixed size so it isn't too squished
	boxSize := fyne.NewSize(450, 300)
	mainLayout := layout.NewGridWrapLayout(boxSize)
	// Put our layout into a container to display it
	mainContainer := fyne.NewContainerWithLayout(mainLayout, fullBox)
	mainWindow.SetContent(mainContainer)
	mainWindow.ShowAndRun()
}
