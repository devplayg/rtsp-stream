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
    return recording;
}

function videosCanPlayFormatter(val, row, idx, field) {
    let id = field.replace("video-", "");
    if (row.date ==="live") {
        let arr = val.split(",");
        if (arr.length < 1) {
            return;
        }
        // console.log(arr);
        let tag = "";
        if (arr[0] === "1") {
            tag += '<a href="#" class="live" data-id="' + id + '">LIVE</a>';
        }
        if (arr[1] === "1") {
            tag += ' <a href="#" class="today" data-id="' + id + '">Today</a>';
        }
        return tag;
    }
    if (val === "1") {
        return '<a href="#" class="play" data-name="won" data-id="' + id + '">a</a>';
    }
}

function streamsCreatedFormatter(val, row, idx) {
    return moment.unix(val).format();
}

function streamsUpdatedFormatter(val, row, idx) {
    return moment.unix(val).format();
}