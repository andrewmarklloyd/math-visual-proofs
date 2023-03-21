# math-visual-proofs-server

github action send message to server to render video
- payload: class/name, file to render? problem with this is filename always changing. Maybe latest committed file?
server acknowleges message
- start render video detached
    - progress somehow?
- check for expected video file to exist
- once exists then upload to spaces
- send notification to user

to think about?
- security
    - what if someone checks in malicious python code?
    - giving access spaces for users
- distributed
    - how can we distribute the render?
- observability
