"use strict";

const app = new Vue({
    el: '#app',
    data: {
        activities: [],
        currentSide: 0,
        tracking: {}
    },
    computed: {
        orderedActivities: function () {
            return _.orderBy(this.activities, ['deviceSide'])
        }
    }
});

const zeiSocket = new WebSocket("ws://localhost:6677/ws");

zeiSocket.onmessage = (data) => {
    let state = JSON.parse(data.data);
    app.activities = state.Activities;
    app.currentSide = state.CurrentSide;
    app.tracking = state.Tracking;
};