function onOffFormatter(tf, idx, row) {
    if (tf) {
        return '<h5><span class="badge badge-success">On</span></h5>';
    }
    return '<h5><span class="badge badge-secondary">Off</span></h5>';
}

function streamsControlFormatter(val, row, idx) {
    let arr = [];
    arr.push('<a href="#" class="delete text-danger " data-id="' + row.id + '"><i class="far fa-times"></i></a>');
    arr.push('<a href="#" class="edit" data-id="' + row.id + '"><i class="far fa-edit"></i></a>');
    if (row.status === Stopped) {
        arr.push('<a href="#" class="start" data-id="' + row.id + '"><i class="far fa-play"></i></a>');
    }

    if (row.status === Started) {
        arr.push('<a href="#" class="stop" data-id="' + row.id + '"><i class="far fa-stop"></i></a>');
    }
    return arr.join(' ');
}
//
// function streamsActiveFormatter(active, row, idx) {
//     let text = "Stopped",
//         css = "btn-default";
//
//     if (row.active) {
//         text = "Running";
//         //css = "btn-primary";
//     }
//
//     return [
//         '<button class="btn btn-xs btn-streams-active ' + css + '" data-id="' + row.id + '">',
//         text,
//         '</button>'
//     ].join("");
// }
//
// function streamsRecordingFormatter(recording, row, idx) {
//     let text = "Stopped",
//         css = "btn-default";
//
//     if (recording) {
//         text = "Recording";
//         //  css = "btn-danger";
//     }
//
//     return [
//         '<button class="btn btn-xs btn-active ' + css + '">',
//         text,
//         '</button>'
//     ].join("");
// }

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

function streamsStatusFormatter(val, row, idx) {
    if (val === Failed) {
        return '<span class="text-danger">' + 'Failed' + '</span>';
    }

    if (val === Stopped) {
        return '<span class="text-muted">Stopped</span>';
    }

    if (val === Stopping) {
        return '<div class="spinner-grow spinner-grow-sm rounded-0 text-warning" role="status"><span class="sr-only">Loading...</span></div> Stopping';
    }

    if (val === Starting) {
        return '<div class="spinner-grow spinner-grow-sm text-info" role="status"><span class="sr-only">Loading...</span></div> Starting';
    }

    if (val === Started) {
        return '<span class="badge badge-danger">Live</span>';
    }

    return val;
}

function streamsUpdatedFormatter(val, row, idx) {
    return moment.unix(val).format();
}

function videosVideoFormatter(val, row, idx) {
    // console.log(row);
    let keys = $.map(row, function(v, k) {
        if (k.startsWith("video-")) {
            return k;
        }
    });
    if (keys.length < 1) return;

    let tags = "";
    $.each(keys, function(i, k) {
        // console.log(k);
        if (row[k] === 1) {
            tags += '<a href="#" class="btn btn-primary btn-xs btn-icon rounded-circle"><i class="fal fa-check"></i></a>';
            return true;
        }
        tags += '<a href="#" class="btn btn-default btn-xs btn-icon rounded-circle">-</a>';
        return true;
    });

    return tags;
    // console.log(keys);

    // var collator = new Intl.Collator(undefined, {numeric: true, sensitivity: 'base'});
    // var myArray = ['1_Document', '11_Document', '2_Document'];
    // console.log(myArray.sort(collator.compare));


    // console.log(row);
    // return row.date;
}

function videosDateFormatter(val, idx, row) {
    let m =  moment(val, 'YYYYMMDD');
    return m.format("ll") + '<span class="small">' + m.format("ddd") + '</span>';
}