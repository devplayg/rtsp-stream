function streamsControlFormatter(val, row, idx) {
    return [
        '<a href="#" class="delete" data-id="' + row.id + '">',
        "Delete",
        '</a>',
        '<br>',

        '<a href="#" class="start" data-id="' + row.id + '">',
        "Start",
        '</a>',
        '<br>',

        '<a href="#" class="stop" data-id="' + row.id + '">',
        "Stop",
        '</a>',
        '<br>',

        '<a href="#" class="edit" data-id="' + row.id + '">',
        "Edit",
        '</a>',

    ].join("");
}

function streamsActiveFormatter(active, row, idx) {
    let text = "Stopped",
        css = "btn-default";

    if (row.active) {
        text = "Running";
        //css = "btn-primary";
    }

    return [
        '<button class="btn btn-xs btn-streams-active ' + css + '" data-id="' + row.id + '">',
        text,
        '</button>'
    ].join("");
}

function streamsRecordingFormatter(recording, row, idx) {
    let text = "Stopped",
        css = "btn-default";

    if (recording) {
        text = "Recording";
        //  css = "btn-danger";
    }

    return [
        '<button class="btn btn-xs btn-active ' + css + '">',
        text,
        '</button>'
    ].join("");
}

function videosCanPlayFormatter(val, row, idx, field) {
    if (val === "1") {
        let id = field.replace("video-", "");
        return '<a href="#" class="play" data-name="won" data-id="' + id + '">a</a>';
    }
}

