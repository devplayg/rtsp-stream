const Failed = -1,
    Stopped = 1,
    Stopping = 2,
    Starting = 3,
    Started = 4;

let colorPallet = [50, 100, 200, 300, 400, 500, 600, 700, 800, 900];

moment.tz.setDefault("Asia/Seoul");

let Streams = function() {
    this.streams = {};
    this.init = function() {
        let c = this;
        $.ajax({
            url: "/streams"
        }).done(function(streams) {
            // console.log(list);
            $.each(streams, function(i, s) {
                c.streams[s.id] = s;
            });
            console.log(c.streams);
        }).fail(function(xhr) {
            Swal.fire({
                type: 'error',
                title: 'Error',
                text: 'Failed to get streams',
            });
            console.error(xhr);
        });
    };

    this.init();
};

streams = new Streams();

// function convertToUserTime(dt) {
//     return moment.tz(dt, systemTz).tz(userTz);
// }

