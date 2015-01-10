import QtQuick 2.0
import Ubuntu.Components 1.1
import Ubuntu.Components.ListItems 1.0 as ListItem

Page {
    id: accountsPage
    title: i18n.tr("Inbox")

    ListView {
        id: listView
        model: threadModel.len
        clip: true
        width: parent.width
        height: parent.height

        delegate: ListItem.Empty {
            width: parent.width
            height: column.height + units.gu(2)

            property var thread: threadModel.getThread(index)

            Column {
                id: column
                spacing: units.gu(0.5)
                anchors {
                    right: parent.right
                    left: parent.left
                    rightMargin: units.gu(2)
                    leftMargin: units.gu(2)
                    verticalCenter: parent.verticalCenter
                }
                Item {
                    width: parent.width
                    height: senderLabel.height

                    Label {
                        id: senderLabel
                        anchors {
                            left: parent.left
//                            right: countLabel.left
                            verticalCenter: parent.verticalCenter
                        }
                        color: "black"
                        fontSize: "medium"
                        elide: Text.ElideRight
                        text: thread.getSendersString()
                    }

                    Label { //TODO: Minimum width
                        id: countLabel
                        anchors {
                            left: senderLabel.right
                            right: dateLabel.left
                            leftMargin: units.gu(0.7)
                            bottom: senderLabel.bottom
                        }
                        fontSize: "small"
                        elide: Text.ElideRight
                        text: thread.messageCount
                    }

                    Label {
                        id: dateLabel
                        anchors {
                            right: parent.right
                            verticalCenter: parent.verticalCenter
                        }

                        fontSize: "small"
                        elide: Text.ElideRight
                        text: thread.getDateString()
                    }
                }
                Label {
                    width: parent.width
                    fontSize: "small"
                    elide: Text.ElideRight
                    text: thread.subject
                }
                Label {
                    width: parent.width
                    fontSize: "small"
                    elide: Text.ElideRight
                    text: thread.snippet
                }
            }
        }
    }
}
