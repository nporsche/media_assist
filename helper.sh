 #!/bin/sh
set -x
./bin/mover -spath=/Volumes/photo/family-upload -dpath=/Volumes/photo -file="*.jpeg;*.JPG;*.MOV;*.mp4;*.png"
./bin/mover -spath=/Volumes/photo/nporsche-upload -dpath=/Volumes/photo -file="*.jpeg;*.JPG;*.MOV;*.mp4;*.png"

find /Volumes/photo/20*/* -name "*(*" | xargs -I[] mv [] /Volumes/photo/duplicated/ 