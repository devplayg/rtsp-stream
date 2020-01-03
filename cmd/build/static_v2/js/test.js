// let player = videojs('player', {
//     controls: true,
//     autoplay: false,
//     preload: 'auto',
//     playbackRates: [0.5, 1, 1.5, 2, 4, 8],
// });
// RemainingTimeDisplay(player, )
// // player.ControlBar.prototype.options_ = {
// //     loadEvent: 'play',
// //     children: ['playToggle', 'volumeMenuButton', 'currentTimeDisplay', 'progressControl', 'liveDisplay', 'durationDisplay', 'customControlSpacer', 'playbackRateMenuButton', 'chaptersButton', 'subtitlesButton', 'captionsButton', 'fullscreenToggle']
// // };
//
//
// player.src({
//     type: "application/x-mpegURL",
//     src: "/videos/1/date/20200102/m3u8"
//
// });
// player.ready(function () {
//     // player.muted(true);
//     player.currentTime(1200);
//     player.play();
// });

var options = {
    controls: true,
    autoplay: false,
    preload: 'auto',
    playbackRates: [0.5, 1, 1.5, 2, 4, 8],
    muted: true,
    sources: [{
        src: "/videos/1/date/20200102/m3u8",
        type: "application/x-mpegURL",
    }],
    controlBar: {
//         bigPlayButton: false,
//         muteToggle: true,
        playToggle: true,
        timeDivider: true,
        currentTimeDisplay: true,
        durationDisplay: true,
        remainingTimeDisplay: true,
        // progressControl: false,
//         fullscreenToggle: false,
//         volumeControl: false,
    },
};
var player = videojs('player', options, function onPlayerReady() {
    videojs.log('Your player is ready!');
    this.play();

    this.on('ended', function() {
        videojs.log('Awww...over so soon?!');
    });
});



//
// var myPlayer = videojs('video1', {
//     autoplay: true,
//     loop: true,
//     controlBar: {
//         bigPlayButton: false,
//         muteToggle: true,
//         playToggle: true,
//         timeDivider: false,
//         currentTimeDisplay: true,
//         durationDisplay: false,
//         remainingTimeDisplay: false,
//         progressControl: false,
//         fullscreenToggle: false,
//         volumeControl: false,
//     }
// });