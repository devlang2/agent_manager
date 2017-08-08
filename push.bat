@echo off
del *.exe 2>-
del *.log 2>-
rmdir temp 2>-
git add *
git commit * -m''
git push

