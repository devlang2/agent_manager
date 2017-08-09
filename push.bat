@echo off
del *.exe 2>nul
del *.log 2>nul
rmdir temp 2>nul
git add *
git commit * -m''
git push

