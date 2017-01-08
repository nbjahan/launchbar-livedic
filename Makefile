SHELL = /bin/bash
DESTDIR = ./dist

RELEASE_BASENAME = Dictionary.Live
BUNDLE_NAME = Dictionary\:\ Define\ \(Live\)
BUNDLE_VERSION = $(shell cat VERSION)
BUNDLE_IDENTIFIER = nbjahan.launchbar.livedic
BUNDLE_ICON = com.apple.Dictionary
AUTHOR = nbjahan
TWITTER = @nbjahan
SLUG = launchbar-livedic
WEBSITE = http://github.com/nbjahan/$(SLUG)
SCRIPT_NAME = dict

LBACTION_PATH = $(DESTDIR)/$(RELEASE_BASENAME).lbaction
RELEASE_FILENAME = $(RELEASE_BASENAME)-$(BUNDLE_VERSION).lbaction
LDFLAGS=

UPDATE_LINK = https://raw.githubusercontent.com/nbjahan/$(SLUG)/master/src/Info.plist
DOWNLOAD_LINK = https://github.com/nbjahan/$(SLUG)/releases/download/v$(BUNDLE_VERSION)/$(RELEASE_FILENAME)
all:
	@$(RM) -rf $(DESTDIR)

	@install -d ${LBACTION_PATH}/Contents/{Resources,Scripts}
	@plutil -replace CFBundleName -string $(BUNDLE_NAME) $(PWD)/src/Info.plist
	@plutil -replace CFBundleVersion -string $(BUNDLE_VERSION) $(PWD)/src/Info.plist
	@plutil -replace CFBundleIdentifier -string $(BUNDLE_IDENTIFIER) $(PWD)/src/Info.plist
	@plutil -replace CFBundleIconFile -string $(BUNDLE_ICON) $(PWD)/src/Info.plist
	@plutil -replace LBDescription.LBAuthor -string $(AUTHOR) $(PWD)/src/Info.plist
	@plutil -replace LBDescription.LBTwitter -string $(TWITTER) $(PWD)/src/Info.plist
	@plutil -replace LBDescription.LBWebsiteURL -string $(WEBSITE) $(PWD)/src/Info.plist
	@plutil -replace LBDescription.LBUpdateURL -string $(UPDATE_LINK) $(PWD)/src/Info.plist
	@plutil -replace LBDescription.LBDownloadURL -string $(DOWNLOAD_LINK) $(PWD)/src/Info.plist
	@plutil -replace LBScripts.LBDefaultScript.LBScriptName -string $(SCRIPT_NAME) $(PWD)/src/Info.plist
	@install -pm 0644 ./src/Info.plist $(LBACTION_PATH)/Contents/
	gb build $(LDFLAGS) $(SCRIPT_NAME) && mv bin/$(SCRIPT_NAME) $(LBACTION_PATH)/Contents/Scripts/
	-@cp -f ./src/*.js $(LBACTION_PATH)/Contents/Scripts/
	-@cp -rf ./resources/* $(LBACTION_PATH)/Contents/Resources/

	@echo "Refreshing the LaunchBar"
	@osascript -e 'run script "tell application \"LaunchBar\" \n repeat with rule in indexing rules \n if name of rule is \"Actions\" then \n update rule \n exit repeat \n end if \n end repeat \n activate \n end tell"'

	@echo "Making a release"
	@install -d $(DESTDIR)/release
	@ditto -ck --keepParent $(LBACTION_PATH)/ $(DESTDIR)/release/$(RELEASE_FILENAME)

dev: LDFLAGS := -ldflags "-X main.InDev true"
dev: all

release: all

.PHONY: all dev
