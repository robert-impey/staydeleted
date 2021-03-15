# staydeleted

[![Build Status](https://travis-ci.com/robert-impey/stay-deleted.svg?branch=master&status=unknown)](https://travis-ci.com/github/robert-impey/stay-deleted)

To make sure that files stay deleted.

This is useful for folders that are synched in both directions with backup disks on a regular schedule.

See https://github.com/robert-impey/generate-synch-scripts

This leads to deleted files being restored from the backup folder.

To get around this, a file or folder can be `marked` for deletion:

`PS C:\foo>staydelelted mark bar.txt`

This creates a special file in a subfolder of the directory containing the file to be deleted.

On a schedule, that folder can be swept clean:

`staydelted sweep C:\foo`

The program will search that folder and its subfolders for the special files and delete the marked files.

If you change your mind, you can mark file to be kept:

`PS C:\foo>staydeleted mark --keep bar.txt`

Note that this does not simply delete the special file.
This would be ineffective as it would be restored itself when the folders were next synchronized.
Instead, the special file contains an instruction to keep the file.
This behaviour relies on the folders being synched with rsync's `--update` option or similar.

Finally, the program cleans up after itself, deleting the special files and their folders after a year.
This should be enough time for all files marked for deletion to be deleted from all backups.
