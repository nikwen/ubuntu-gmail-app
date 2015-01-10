import QtQuick 2.0
import Ubuntu.Components 1.1
import Ubuntu.OnlineAccounts 0.1
import Ubuntu.OnlineAccounts.Client 0.1
import Ubuntu.Components.ListItems 1.0 as ListItem

MainView {
    applicationName: "com.ubuntu.developer.nikwen.ubuntu-gmail-app"

    useDeprecatedToolbar: false

    width: units.gu(100)
    height: units.gu(75)

    PageStack {
        id: pageStack

        Component.onCompleted: pageStack.push(accountsPage)
    }

    Page {
        id: accountsPage
        visible: false
        title: i18n.tr("GMail app proof of concept")

        property bool finishedInitialSetup: false

        head.actions: [
            Action {
                iconName: "add"
                text: i18n.tr("Add account")
                onTriggered: {
                    accountsPage.finishedInitialSetup = false //TODO: Dialog!
                    setup.exec()
                }
            }
        ]

        AccountServiceModel {
            id: accounts
            includeDisabled: false
            applicationId: "com.ubuntu.developer.nikwen.ubuntu-gmail-app_ubuntu-gmail-app"
        }

        Setup {
            id: setup
            applicationId: accounts.applicationId
            providerId: "google"

            onFinished: {
                accountsPage.finishedInitialSetup = true
            }
        }

        Component.onCompleted: {
            if (accounts.count === 0) {
                setup.exec()
            } else {
                finishedInitialSetup = true
            }
        }

        ActivityIndicator {
            anchors.centerIn: listView
            visible: !accountsPage.finishedInitialSetup
            running: visible
        }

        ListView {
            id: listView
            model: accounts
            clip: true
            width: parent.width
            height: parent.height

            delegate: ListItem.Standard {
                text: model.displayName

                AccountService {
                    id: accountService
                    objectHandle: model.accountServiceHandle
                    onAuthenticated: {
                        var params = accountService.authData.parameters
                        gmailBackend.instantiateGmailService(params.ClientId, params.ClientSecret, params.Scope[0], reply.AccessToken)
                        gmailBackend.listThreadsAsynchronously()
                        pageStack.push(Qt.resolvedUrl("Inbox.qml"))
                    }
                    onAuthenticationError: { console.log("Authentication failed, code " + error.code) }
                }

                onClicked: accountService.authenticate({})
            }
        }

        Label {
            visible: accounts.count === 0 && accountsPage.finishedInitialSetup
            anchors {
                left: parent.left
                right: parent.right
                verticalCenter: listView.verticalCenter
                margins: units.gu(3)
            }
            fontSize: "large"
            wrapMode: Text.Wrap
            horizontalAlignment: Text.AlignHCenter

            text: i18n.tr("Please an account by clicking on the plus icon in the header.")
        }
    }
}

