# btdash-packager
The goal of this project is to create a DASH like video delivery techinque through a BitTorrent network. This tool encodes 
a given video file in to a DASH formatted MP4, analyzes the file, and then creates a torrent bundle with additional AV/segment information 
embed in the torrent file. This torrent file can then be used by supported players ([btdash-http-player](https://github.vimeows.com/thomas/btdash-http-player)) to play back the video file with full seek control.

## Usage
```
Usage of ./btdash-packager [inputFile]:
  -force-transcode
    	Should the input file be transmuxed regardless of if it has to
  -json
    	Output JSON formatted manifest
  -output string
    	Output directory for torrent and video files (default "./")
```

For example `./btdash-packager in.mkv -json` outputs three files in to your current directory
- `out.mp4` is the DASH encoded video rendition of `in.mkv` 
- `out.torrent` links to out.mp4 and has a playback manifest bencoded in its metadata field
- `out.json` contains the playback manifest in a JSON encoded form

#### MP4 Boxdump
```
$ boxdumper out.mp4
[File]
    size = 393685999
    [ftyp: File Type Box]
    [moov: Movie Box]
    [sidx: Segment Index Box]
    [sidx: Segment Index Box]
    [moof: Movie Fragment Box]
    [mdat: Media Data Box]
    ...
    [sidx: Segment Index Box]
    [sidx: Segment Index Box]
    [moof: Movie Fragment Box]
    [mdat: Media Data Box]
```
