function streamsControlFormatter(val, row, idx) {
    return [
        '<a href="#" class="delete data-id="' + row.id + '">',
        "Delete",
        '</a>',
        '<button class="btn btn-default btn-xs start data-id="' + row.id + '">',
        "Start",
        '</button>',
        '<button class="btn btn-default btn-xs stop data-id="' + row.id + '">',
        "Stop",
        '</button>',
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