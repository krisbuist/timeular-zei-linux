"use strict";

const app = new Vue({
    el: '#app',
    data: {
        Activities: [],
        CurrentSide: 0,
        Tracking: {}
    }
});

const zeiSocket = new WebSocket("ws://localhost:6677/ws");

zeiSocket.onmessage = (data) => {
    let state = JSON.parse(data.data);
    app.Activities = state.Activities;
    app.CurrentSide = state.CurrentSide;
    app.Tracking = state.Tracking;
};