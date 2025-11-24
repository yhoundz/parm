Set WindowBar Colorful
Set Theme "TokyoNight"
Set LineHeight 1.6
Require parm
Output assets/demo.gif

Set FontSize 16
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

Type "parm install jesseduffield/lazygit"
Sleep 0.1s
Type "@v0.55.0"
Enter
Wait />$/

Type "parm install sxyazi/yazi --pre-release --strict"
Sleep 0.1s
Enter
Wait />$/

Type "parm install charmbracelet/crush --asset crush_0.18.5_Linux_x86_64.tar.gz"
Sleep 0.1s
Enter
Wait />$/

Type "clear"
Sleep 0.1s
Enter
Wait />$/

Type "parm list"
Sleep 0.1s
Enter
Wait />$/

Type "parm info sxyazi/yazi"
Sleep 0.1s
Enter
Wait />$/

Type "parm remove sxyazi/yazi"
Sleep 0.1s
Enter
Wait />$/

Type "which yazi"
Sleep 0.1s
Enter
Wait />$/

Type "parm list"
Sleep 0.1s
Enter
Wait />$/

Hide
Type "parm remove burntsushi/ripgrep charmbracelet/crush jesseduffield/lazygit"
Enter
