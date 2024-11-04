function modifyUrl(url) {
    let origin = window.location.origin

    return url.replace('https://chatgpt.com', origin)
        .replace('https://ab.chatgpt.com', origin + '/ab')
        .replace('https://cdn.oaistatic.com', origin)
        .replace('wss://webrtc.chatgpt.com', origin.replace('http', 'ws') + '/webrtc')
}

window.fetch = new Proxy(window.fetch, {
    apply: function(target, thisArg, argumentsList) {
        argumentsList[0] = modifyUrl(argumentsList[0])
        if (argumentsList[0].indexOf('/ab/v1/rgstr') !== -1) {
            const body = argumentsList[1] || argumentsList[1].body
            if (body) {
                argumentsList[1].body = argumentsList[1].body.replaceAll(window.location.host, 'chatgpt.com')
            }
        }
        return target.apply(thisArg, argumentsList).then(response => {
            if (response.status >= 401) {
                return target.apply(thisArg, argumentsList)
            }
            return response
        });
    }
});

window.WebSocket = new Proxy(window.WebSocket, {
    construct: function(target, argumentsList) {
        argumentsList[0] = modifyUrl(argumentsList[0])
        return new target(...argumentsList);
    }
});


const originalSrcDescriptor = Object.getOwnPropertyDescriptor(HTMLImageElement.prototype, 'src');
Object.defineProperty(HTMLImageElement.prototype, 'src', {
    set(value) {
        const modifiedUrl = modifyUrl(value);
        originalSrcDescriptor.set.call(this, modifiedUrl);
    },
    get() {
        return originalSrcDescriptor.get.call(this);
    }
});


