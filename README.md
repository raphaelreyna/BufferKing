## BufferKing

Cache it your way!
Record your computers audio output to an audio library.


### About

BufferKing listens for 'Play/Pause/Next/Previous' notifications from any music player using DBus. When it detects that a new track is being played it will start recording it.


### Limitations

BufferKing doesn't handle pausing very well (at all) right now.
If a track is being recorded and is paused or skipped it is discarded by default.
This behavior can be changed using the `-P, --keep-paused` and `-S, --keep-skipped` flags if you wish to keep partial recordings to stitch together later or whatever.


### Contributing

Contributions are very welcome :)
