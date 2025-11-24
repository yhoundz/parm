Set WindowBar Colorful
Set Theme "catppuccin-mocha"
Set LineHeight 1.6
Require parm
Output assets/demo.gif

Set FontSize 24
Set TypingSpeed 60ms
Set Width 1440
Set Height 960

Hide
Type "export CI=false TERM=xterm-256color && clear"
Enter
Show

Type "parm install burntsushi/ripgrep"
Sleep 0.1s
Enter
Wait />$/
Sleep 0.5s

Type "parm install jesseduffield/lazygit"
Sleep 0.1s
Type "@v0.55.0"
Enter
Wait />$/
Sleep 0.5s

Type "parm install sxyazi/yazi --pre-release --strict"
Sleep 0.1s
Enter
Wait />$/
Sleep 0.5s

Type "parm install charmbracelet/crush --asset crush_0.18.5_Linux_x86_64.tar.gz"
Sleep 0.1s
Enter
Wait />$/
Sleep 2s

Type "clear"
Sleep 0.1s
Enter
Wait />$/
Sleep 1s

Type "parm list"
Sleep 0.1s
Enter
Wait />$/
Sleep 2s

Type "parm info sxyazi/yazi"
Sleep 0.1s
Enter
Wait />$/
Sleep 2s

Type "parm remove sxyazi/yazi"
Sleep 0.1s
Enter
Wait />$/
Sleep 2s

Type "parm list"
Sleep 0.1s
Enter
Wait />$/
Sleep 4s

Hide
Type "parm remove burntsushi/ripgrep charmbracelet/crush jesseduffield/lazygit"
Enter
