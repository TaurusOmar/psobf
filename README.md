# Powershell Obfuscator
Tool for obfuscating PowerShell scripts written in Go. The main objective of this program is to obfuscate PowerShell code to make its analysis and detection more difficult.
The script offers 5 levels of obfuscation, from basic obfuscation to script fragmentation. This allows users to tailor the obfuscation level to their specific needs.

```powershell 

./psobf -h

	██████╗ ███████╗ ██████╗ ██████╗ ███████╗
	██╔══██╗██╔════╝██╔═══██╗██╔══██╗██╔════╝
	██████╔╝███████╗██║   ██║██████╔╝█████╗
	██╔═══╝ ╚════██║██║   ██║██╔══██╗██╔══╝
	██║     ███████║╚██████╔╝██████╔╝██║
	╚═╝     ╚══════╝ ╚═════╝ ╚═════╝ ╚═╝
	@TaurusOmar
	v.1.0											 	
	
Usage: ./obfuscator -i <inputFile> -o <outputFile> -level <1|2|3|4|5>
Options:
  -i string
    	Name of the PowerShell script file.
  -level int
    	Obfuscation level (1 to 5). (default 1)
  -o string
    	Name of the output file for the obfuscated script. (default "obfuscated.ps1")

Obfuscation levels:
  1: Basic obfuscation by splitting the script into individual characters.
  2: Base64 encoding of the script.
  3: Alternative Base64 encoding with a different PowerShell decoding method.
  4: Compression and Base64 encoding of the script will be decoded and decompressed at runtime.
  5: Fragmentation of the script into multiple parts and reconstruction at runtime.
```

## Features:
- Obfuscation Levels: Four levels of obfuscation, each more complex than the previous one.
    - Level 1 obfuscation by splitting the script into individual characters.
    - Level 2 Base64 encoding of the script.
    - Level 3 Alternative Base64 encoding with a different PowerShell decoding method.
    - Level 4 Compression and Base64 encoding of the script will be decoded and decompressed at runtime.
    - Level 5 Fragmentation of the script into multiple parts and reconstruction at runtime.
- Compression and Encoding: Level 4 includes script compression before encoding it in base64.
- Variable Obfuscation: A function was added to obfuscate the names of variables in the PowerShell script.
- Random String Generation: Random strings are generated for variable name obfuscation.

![image](https://github.com/TaurusOmar/psobf/blob/main/psobf.gif?raw=true)

## Install
```
go install github.com/TaurusOmar/psobf@latest
```

## Example of Obfuscation Levels
The obfuscation levels are divided into 5 options. First, you need to have a PowerShell file that you want to obfuscate. Let's assume you have a file named `script.ps1` with the following content:

```powershell
Write-Host "Hello, World!"
```

### Level 1: Basic Obfuscation

Run the script with level 1 obfuscation.
```powershell
./obfuscator -i script.ps1 -o obfuscated_level1.ps1 -level 1
```
This will generate a file named `obfuscated_level1.ps1` with the obfuscated content. The result will be a version of your script where each character is separated by commas and combined at runtime.<br>
**Result (level 1)**
```powershell
$obfuscated = $([char[]]("`W`,`r`,`i`,`t`,`e`,`-`,`H`,`o`,`s`,`t`,` `,`"`,`H`,`e`,`l`,`l`,`o`,`,` `,`W`,`o`,`r`,`l`,`d`,`!`,`"`") -join ''); Invoke-Expression $obfuscated
```

### Level 2: Base64 Encoding
Run the script with level 2 obfuscation:
```powershell
./obfuscator -i script.ps1 -o obfuscated_level2.ps1 -level 2
```
This will generate a file named `obfuscated_level2.ps1` with the content encoded in base64. When executing this script, it will be decoded and run at runtime.<br>
**Result (level 2)**
```powershell
$obfuscated = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('V3JpdGUtSG9zdCAiSGVsbG8sIFdvcmxkISI=')); Invoke-Expression $obfuscated
```

### Level 3: Alternative Base64 Encoding
Execute the script with level 3 obfuscation:
```powershell
./obfuscator -i script.ps1 -o obfuscated_level3.ps1 -level 3
```
This level uses a slightly different form of base64 encoding and decoding in PowerShell, adding an additional layer of obfuscation.<br>
**Result (level 3)**
```powershell
$e = [System.Convert]::FromBase64String('V3JpdGUtSG9zdCAiSGVsbG8sIFdvcmxkISI='); $obfuscated = [System.Text.Encoding]::UTF8.GetString($e); Invoke-Expression $obfuscated
```

### Level 4: Compression and Base64 Encoding
Execute the script with level 4 obfuscation:
```powershell
./obfuscator -i script.ps1 -o obfuscated_level4.ps1 -level 4
```
This level compresses the script before encoding it in base64, making analysis more complicated. The result will be decoded and decompressed at runtime.<br>
**Result (level 4)**
```powershell
$compressed = 'H4sIAAAAAAAAC+NIzcnJVyjPL8pJUQQAlRmFGwwAAAA='; $bytes = [System.Convert]::FromBase64String($compressed); $stream = New-Object IO.MemoryStream(, $bytes); $decompressed = New-Object IO.Compression.GzipStream($stream, [IO.Compression.CompressionMode]::Decompress); $reader = New-Object IO.StreamReader($decompressed); $obfuscated = $reader.ReadToEnd(); Invoke-Expression $obfuscated
```

### Level 5: Script Fragmentation
Run the script with level 5 obfuscation:
```powershell
./obfuscator -i script.ps1 -o obfuscated_level5.ps1 -level 5
```
This level fragments the script into multiple parts and reconstructs it at runtime.<br>
**Result (level 5)**
```powershell
$fragments = @(
'Write-', 
'Output "', 
'Hello,', 
' Wo', 
'rld!', 
'"'
); 
$script = $fragments -join ''; 
Invoke-Expression $script
```

This program is provided for educational and research purposes. It should not be used for malicious activities.
