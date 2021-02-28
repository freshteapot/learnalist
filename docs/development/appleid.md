# Apple ID

# Thoughts
- Do I need challenge via state query parameter?
- Setting up the app to use sign in, is enough for apple.
- It might be possible (hope!) to reuse just one clientID + secret via android as it will use the web setup
- Merging accounts is going to be a pain due to the event nature
    - event to say it happened
    - manually move everything in multiple databases :P
    - event to trigger deletion of old uuid
    - logout all sessions
        - wonder how the apps handle this
    - rebuild users:
        - challenges
        - lists
    - double check challenges as I do this. Should be happy
