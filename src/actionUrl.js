function runWithURL(url , details) {
  LaunchBar.log(JSON.stringify(details));
  var path = details.path.replace('/nbjahan.launchbar.livedic/', '');
  switch(path) {
    case 'lookup':
      var q = details.query;
      if (q && q.length > 0) {
        q = JSON.stringify(q).slice(1,-1);
        LaunchBar.performAction("Dictionary: Define (Live)", q);
      }
  }
}