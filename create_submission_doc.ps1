$word = New-Object -ComObject Word.Application
$word.Visible = $false
$doc = $word.Documents.Add()
$selection = $word.Selection

# Title
$selection.Font.Name = "Segoe UI"
$selection.Font.Size = 20
$selection.Font.Bold = 1
$selection.Font.Color = 0x2A1DE1 # Crimson Red
$selection.TypeText("Dream11 Assignment Submission`n")

# Subtitle / Candidate
$selection.Font.Size = 13
$selection.Font.Bold = 1
$selection.Font.Color = 0x333333
$selection.TypeText("Forward Deployed Engineer Intern`n")
$selection.Font.Bold = 0
$selection.Font.Size = 11
$selection.TypeText("Candidate: Soham Balwant`n`n")

# Quick Reference Box / Metadata
$selection.Font.Bold = 1
$selection.Font.Size = 11
$selection.Font.Color = 0x000000
$selection.TypeText("Artifact Link: ")
$selection.Font.Bold = 0
$selection.Font.Color = 0xCC0000
$selection.TypeText("https://dream11-roan.vercel.app`n")

$selection.Font.Bold = 1
$selection.Font.Color = 0x000000
$selection.TypeText("GitHub Repository: ")
$selection.Font.Bold = 0
$selection.Font.Color = 0xCC0000
$selection.TypeText("https://github.com/SohamBalwant/Dream11`n`n")

$selection.Font.Bold = 1
$selection.Font.Color = 0x000000
$selection.TypeText("Demo Login Credentials:`n")
$selection.Font.Bold = 0
$selection.TypeText("• Email: cricketer@dream11.com`n")
$selection.TypeText("• Password: Password@123`n")
$selection.Font.Italic = 1
$selection.Font.Size = 10
$selection.Font.Color = 0x666666
$selection.TypeText("(A one-click 'Auto-fill' button is also available in the top-right corner of the login screen)`n`n")
$selection.Font.Italic = 0
$selection.Font.Size = 11
$selection.Font.Color = 0x000000

# Question 1
$selection.Font.Bold = 1
$selection.Font.Size = 12
$selection.Font.Color = 0x2A1DE1
$selection.TypeText("1. What fan problem are you solving?`n")
$selection.Font.Bold = 0
$selection.Font.Size = 11
$selection.Font.Color = 0x222222
$selection.TypeText("When friends watch live sports together remotely, they juggle three disjointed apps: the video stream, WhatsApp banter, and Dream11. Fans miss key moments switching tabs and struggle to immediately calculate how a sudden boundary or wicket impacts their fantasy leaderboard versus their friends in real time.`n`n")

# Question 2
$selection.Font.Bold = 1
$selection.Font.Size = 12
$selection.Font.Color = 0x2A1DE1
$selection.TypeText("2. How does your solution help?`n")
$selection.Font.Bold = 0
$selection.Font.Size = 11
$selection.Font.Color = 0x222222
$selection.TypeText("We built 'Match Buddy', an AI-powered live fantasy companion embedded directly inside a synchronized watch party. Fans paste any live stream URL (like YouTube) and watch in sync. After every ball, Match Buddy automatically calculates exact fantasy point deltas (+6 for sixes, +25 for wickets, captain multipliers), announces lead swings in chat, tracks head-to-head rankings, and runs instant 'What-If' projections (projecting leaderboard shifts if a 6, 4, or wicket drops).`n`n")

# Question 3
$selection.Font.Bold = 1
$selection.Font.Size = 12
$selection.Font.Color = 0x2A1DE1
$selection.TypeText("3. Which AI tools did you use?`n")
$selection.Font.Bold = 0
$selection.Font.Size = 11
$selection.Font.Color = 0x222222
$selection.TypeText("Google Gemini API for natural-language match commentary, witty point-swing summaries, and conversational companion Q&A ('Who is winning?', 'What if a boundary is scored?'), paired with an event-driven Go and WebSocket state engine.`n")

# Save paths
$docxDesktop = "C:\Users\91932\Desktop\Soham_Balwant_Dream11_Submission.docx"
$pdfDesktop = "C:\Users\91932\Desktop\Soham_Balwant_Dream11_Submission.pdf"
$docxRepo = "d:\inox\Soham_Balwant_Dream11_Submission.docx"
$pdfRepo = "d:\inox\Soham_Balwant_Dream11_Submission.pdf"

$doc.SaveAs([ref]$docxDesktop, [ref]16)
$doc.SaveAs([ref]$pdfDesktop, [ref]17)
$doc.SaveAs([ref]$docxRepo, [ref]16)
$doc.SaveAs([ref]$pdfRepo, [ref]17)

$doc.Close()
$word.Quit()
[System.Runtime.Interopservices.Marshal]::ReleaseComObject($word) | Out-Null
Write-Output "GENERATED_SUCCESSFULLY"
