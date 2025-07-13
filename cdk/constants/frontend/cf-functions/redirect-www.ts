async function handler(event) {
  var request = event.request;
  var headers = request.headers;
  var host = headers.host.value;

  if (host.startsWith('www.')) {
    var newHost = host.slice(4);
    var response = {
      statusCode: 301,
      statusDescription: 'Moved Permanently',
      headers:
        { "location": { value: `https://${newHost}${request.uri}` } }
    }

    return response;
  }
  return request;
}

