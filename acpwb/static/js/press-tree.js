/* Answer-tree figure for press releases: a sunburst of one test attempt's answer tree,
 * coloured by when each page was first requested. Markup comes from the `tree` block in
 * press_release_detail.html; data is the JSON named by data-src. */
(function () {
  'use strict';
  var NAVY = '#0A1628', GOLD = '#C9A84C', EMPTY = '#e9e4d3', INK = '#222', RED = '#8b1e1e', PASS = '#1e3560';
  var DEPTH = 6;

  var STOPS = [[242, 223, 160], [201, 168, 76], [93, 111, 148], [10, 22, 40]];  // pale gold, gold, slate, navy

  function ramp(f) {  // early = pale gold, late = navy; same stops as tools/generate_press_tree.py
    var x = Math.max(0, Math.min(1, f)) * (STOPS.length - 1), i = Math.min(Math.floor(x), STOPS.length - 2), k = x - i;
    return 'rgb(' + STOPS[i].map(function (v, n) { return Math.round(v + (STOPS[i + 1][n] - v) * k); }).join(',') + ')';
  }

  function clock(t0, m) {
    var p = t0.match(/(\d+)-(\d+) (\d+):(\d+)/);
    var d = new Date(Date.UTC(2026, +p[1] - 1, +p[2], +p[3], +p[4]) + m * 6e4);
    var z = function (n) { return (n < 10 ? '0' : '') + n; };
    return 'Oct ' + d.getUTCDate() + ', ' + z(d.getUTCHours()) + ':' + z(d.getUTCMinutes()) + ' UTC';
  }

  function Tree(root, data) {
    this.root = root; this.D = data;
    this.att = +root.getAttribute('data-attempt') || 3;
    this.winner = root.getAttribute('data-winner') || '';
    this.winAtt = +root.getAttribute('data-winner-attempt') || 0;
    this.cut = data.span; this.hover = null; this.timer = null;
    this.build();
  }

  Tree.prototype.build = function () {
    var self = this, D = this.D;
    this.root.innerHTML =
      '<div class="pt-bar"><div class="pt-seg" role="group" aria-label="Attempt"></div>' +
      '<button type="button" class="pt-play">Play</button></div>' +
      '<div class="pt-time"><input type="range" min="0" max="' + D.span + '" value="' + D.span + '" aria-label="Time since the first request">' +
      '<span class="pt-clock"></span></div>' +
      '<div class="pt-stage"><div class="pt-cv"><canvas aria-label="Sunburst of the answer tree for the selected attempt"></canvas></div>' +
      '<div class="pt-side"><div><div class="pt-k">Result pages fetched</div><div class="pt-v pt-n1"></div></div>' +
      '<div><div class="pt-k">Passing pages found</div><div class="pt-v pt-n2"></div></div>' +
      '<div><div class="pt-k">First requested</div><div class="pt-ramp"></div><div class="pt-lg"><span class="pt-l0"></span><span class="pt-l1"></span></div></div>' +
      '<div class="pt-read">Hover or tap a wedge.</div></div></div>';
    var q = function (s) { return self.root.querySelector(s); };
    this.cv = q('canvas'); this.ctx = this.cv.getContext('2d');
    this.slider = q('input'); this.clockEl = q('.pt-clock'); this.read = q('.pt-read');
    this.playBtn = q('.pt-play');
    var seg = q('.pt-seg');
    Object.keys(D.att).forEach(function (a) {
      var b = document.createElement('button'); b.type = 'button'; b.textContent = a; b.setAttribute('data-a', a);
      b.setAttribute('aria-label', 'Attempt ' + a); b.addEventListener('click', function () { self.setAttempt(+a); });
      seg.appendChild(b);
    });
    var stops = []; for (var i = 0; i <= 10; i++) stops.push(ramp(i / 10));
    q('.pt-ramp').style.background = 'linear-gradient(90deg,' + stops.join(',') + ')';
    q('.pt-l0').textContent = clock(D.t0, 0); q('.pt-l1').textContent = clock(D.t0, D.span);
    this.slider.addEventListener('input', function () { self.stop(); self.cut = +self.slider.value; self.draw(); });
    this.playBtn.addEventListener('click', function () { self.timer ? self.stop() : self.play(); });
    ['pointermove', 'pointerdown'].forEach(function (ev) { self.cv.addEventListener(ev, function (e) { self.pick(e); }); });
    this.cv.addEventListener('pointerleave', function () { self.hover = null; self.describe(); });
    window.addEventListener('resize', function () { self.size(); });
    this.root.classList.add('pt-live');
    this.setAttempt(this.att);
    this.size();
  };

  Tree.prototype.layout = function (a) {
    var L = this.D.att[a].leaves, N = L.length, levels = [];
    for (var d = 1; d <= DEPTH; d++) {
      var arr = [], cur = null;
      for (var i = 0; i < N; i++) {
        var s = L[i].slice(0, d);
        if (!cur || cur.s !== s) { cur = { s: s, i0: i, i1: i + 1 }; arr.push(cur); } else cur.i1 = i + 1;
      }
      arr.forEach(function (n) { n.a0 = n.i0 / N * 2 * Math.PI - Math.PI / 2; n.a1 = n.i1 / N * 2 * Math.PI - Math.PI / 2; });
      levels.push(arr);
    }
    return { levels: levels, N: N };
  };

  Tree.prototype.setAttempt = function (a) {
    this.att = a; this.lay = this.layout(a); this.hover = null;
    [].forEach.call(this.root.querySelectorAll('.pt-seg button'), function (b) {
      b.setAttribute('aria-pressed', +b.getAttribute('data-a') === a);
    });
    this.describe();
  };

  Tree.prototype.size = function () {
    var w = Math.round(this.cv.parentNode.getBoundingClientRect().width) || 480;
    this.dpr = window.devicePixelRatio || 1; this.W = w;
    this.cv.width = w * this.dpr; this.cv.height = w * this.dpr; this.draw();
  };

  Tree.prototype.geom = function () {
    var W = this.W, R = W / 2 - 16, r0 = W * 0.06;
    return { cx: W / 2, R: R, r0: r0, th: (R - r0) / DEPTH };
  };

  Tree.prototype.wedge = function (g, d, n) {
    var c = this.ctx, ri = g.r0 + (d - 1) * g.th, ro = ri + g.th - 1;
    c.beginPath(); c.arc(g.cx, g.cx, ro, n.a0, n.a1); c.arc(g.cx, g.cx, ri, n.a1, n.a0, true); c.closePath();
  };

  Tree.prototype.draw = function () {
    if (!this.lay) return;
    var c = this.ctx, g = this.geom(), A = this.D.att[this.att], span = this.D.span, cut = this.cut, W = this.W;
    c.setTransform(this.dpr, 0, 0, this.dpr, 0, 0); c.clearRect(0, 0, W, W);
    var seen = 0, pass = 0, self = this;
    this.lay.levels.forEach(function (lv, di) {
      lv.forEach(function (n) {
        var t = A.seen[n.s], vis = t !== undefined && t <= cut;
        self.wedge(g, di + 1, n); c.fillStyle = vis ? ramp(t / span) : EMPTY; c.fill();
        if (di + 1 === DEPTH && vis) { seen++; if (A.pass[n.s] !== undefined) pass++; }
      });
    });
    this.lay.levels[DEPTH - 1].forEach(function (n) {
      if (A.pass[n.s] === undefined) return;
      var t = A.seen[n.s], vis = t !== undefined && t <= cut, win = self.isWinner(n.s);
      if (!vis) return;
      var a = (n.a0 + n.a1) / 2, len = win ? 14 : 8;
      c.strokeStyle = win ? RED : PASS; c.lineWidth = win ? 3 : 1.6; c.beginPath();
      c.moveTo(g.cx + Math.cos(a) * (g.R + 2), g.cx + Math.sin(a) * (g.R + 2));
      c.lineTo(g.cx + Math.cos(a) * (g.R + len), g.cx + Math.sin(a) * (g.R + len)); c.stroke();
    });
    if (this.winAtt === this.att && this.winner) {
      c.strokeStyle = RED; c.lineWidth = 2;
      for (var d = 1; d <= DEPTH; d++) {
        var s = this.winner.slice(0, d), n = this.lay.levels[d - 1].filter(function (x) { return x.s === s; })[0];
        var t = n && A.seen[s]; if (!n || t === undefined || t > cut) continue;
        this.wedge(g, d, n); c.stroke();
      }
    }
    if (this.hover) { c.strokeStyle = GOLD; c.lineWidth = 2.5; this.wedge(g, this.hover.d, this.hover.n); c.stroke(); }
    this.root.querySelector('.pt-n1').textContent = seen.toLocaleString() + ' of ' + this.lay.N.toLocaleString();
    this.root.querySelector('.pt-n2').textContent = pass + ' of ' + Object.keys(A.pass).length;
    this.clockEl.textContent = clock(this.D.t0, cut);
  };

  Tree.prototype.isWinner = function (s) { return this.winAtt === this.att && s === this.winner; };

  Tree.prototype.pick = function (e) {
    var r = this.cv.getBoundingClientRect(), g = this.geom();
    var x = (e.clientX - r.left) / r.width * this.W - this.W / 2, y = (e.clientY - r.top) / r.width * this.W - this.W / 2;
    var rad = Math.sqrt(x * x + y * y);
    if (rad < g.r0 || rad > g.R) { this.hover = null; return this.describe(); }
    var d = Math.min(DEPTH, Math.floor((rad - g.r0) / g.th) + 1);
    var ang = Math.atan2(y, x) + Math.PI / 2; if (ang < 0) ang += 2 * Math.PI;
    var i0 = Math.floor(ang / (2 * Math.PI) * this.lay.N);
    var n = this.lay.levels[d - 1].filter(function (q) { return q.i0 <= i0 && i0 < q.i1; })[0];
    this.hover = n ? { n: n, d: d } : null; this.describe();
  };

  Tree.prototype.describe = function () {
    var h = this.hover, el = this.read;
    if (!h) { el.textContent = 'Hover or tap a wedge.'; return this.draw(); }
    var A = this.D.att[this.att], n = h.n, t = A.seen[n.s], sc = A.pass[n.s], cnt = n.i1 - n.i0;
    var html = 'Attempt ' + this.att + ' · answers <b>' + n.s + '</b><br>' + 'after ' + h.d + ' of ' + DEPTH + ' answers · ' + cnt + ' result page' + (cnt > 1 ? 's' : '') + ' below<br>' +
      (t === undefined ? 'not requested before the certificate' : 'first requested ' + clock(this.D.t0, t));
    if (h.d === DEPTH && sc !== undefined) html += '<br>passes (' + sc + '%)';
    if (this.isWinner(n.s)) html += '<br><b style="color:' + RED + '">earned the certificate</b>';
    el.innerHTML = html; this.draw();
  };

  Tree.prototype.play = function () {
    var self = this; if (this.cut >= this.D.span) this.cut = 0;
    this.playBtn.textContent = 'Pause';
    this.timer = setInterval(function () {
      self.cut = Math.min(self.D.span, self.cut + 8); self.slider.value = self.cut; self.draw();
      if (self.cut >= self.D.span) self.stop();
    }, 40);
  };
  Tree.prototype.stop = function () { if (this.timer) { clearInterval(this.timer); this.timer = null; } this.playBtn.textContent = 'Play'; };

  function init() {
    [].forEach.call(document.querySelectorAll('.press-tree[data-src]'), function (el) {
      if (el.getAttribute('data-ready')) return;
      el.setAttribute('data-ready', '1');
      var x = new XMLHttpRequest(); x.open('GET', el.getAttribute('data-src'));
      x.onload = function () { if (x.status === 200) { try { new Tree(el, JSON.parse(x.responseText)); } catch (e) { /* keep the fallback image */ } } };
      x.send();
    });
  }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init); else init();
})();
