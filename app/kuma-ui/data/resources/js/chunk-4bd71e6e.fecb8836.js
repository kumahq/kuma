(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-4bd71e6e"],{"1af6":function(t,e,a){var n=a("63b6");n(n.S,"Array",{isArray:a("9003")})},2055:function(t,e,a){},"20fd":function(t,e,a){"use strict";var n=a("d9f6"),r=a("aebd");t.exports=function(t,e,a){e in t?n.f(t,e,r(0,a)):t[e]=a}},2226:function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"overview"},[n("page-header",{attrs:{noflex:""}},[n("h2",{staticClass:"xxl"},[t._v("\n      "+t._s(this.$route.meta.title)+"\n    ")])]),n("MetricGrid",{attrs:{metrics:t.overviewMetrics}}),t.tableData.data.length?n("KTable",{attrs:{"has-hover":"",options:t.tableData},scopedSlots:t._u([{key:"actions",fn:function(e){var a=e.row;return[n("router-link",{attrs:{to:{name:"mesh-overview",params:{mesh:a.name}}}},[t._v("\n        View Entity\n      ")])]}}],null,!1,847071690)}):n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[n("div",{staticClass:"card-icon mb-3"},[n("img",{attrs:{src:a("a448")}})]),t._v("\n      No meshes found!\n    ")])],2)],1)},r=[],o=a("75fc"),i=a("43c3"),s=a("be10"),c={name:"Overview",metaInfo:function(){return{title:this.$route.meta.title}},components:{MetricGrid:s["a"],PageHeader:i["a"]},data:function(){return{tableData:{headers:[{label:"Name",key:"name"},{label:"Type",key:"type"},{key:"actions",hideLabel:!0}],data:[]}}},computed:{overviewMetrics:function(){return[{metric:"Total Number of Meshes",value:this.$store.state.totalMeshCount},{metric:"Total Number of Dataplanes",value:this.$store.state.totalDataplaneCount}]}},watch:{$route:function(t,e){this.bootstrap()}},beforeMount:function(){this.bootstrap()},methods:{bootstrap:function(){var t=this;this.$store.dispatch("getMeshTotalCount"),this.$store.dispatch("getDataplaneFromMeshTotalCount",this.$route.params.mesh),this.$store.dispatch("getDataplaneTotalCount");var e=function(){return t.$api.getAllMeshes().then((function(e){var a=e.items;a&&a.length&&(t.tableData.data=Object(o["a"])(a))})).catch((function(t){console.error(t)}))};e()}}},u=c,l=(a("c4a3"),a("2877")),f=Object(l["a"])(u,n,r,!1,null,null,null);e["default"]=f.exports},3846:function(t,e,a){a("9e1e")&&"g"!=/./g.flags&&a("86cc").f(RegExp.prototype,"flags",{configurable:!0,get:a("0bfb")})},"43c3":function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("header",{staticClass:"page-header",class:{"flex justify-between items-center my-6":!t.noflex}},[t._t("default")],2)},r=[],o={props:{noflex:{type:Boolean,default:!1}}},i=o,s=(a("e234"),a("2877")),c=Object(s["a"])(i,n,r,!1,null,null,null);e["a"]=c.exports},"549b":function(t,e,a){"use strict";var n=a("d864"),r=a("63b6"),o=a("241e"),i=a("b0dc"),s=a("3702"),c=a("b447"),u=a("20fd"),l=a("7cd6");r(r.S+r.F*!a("4ee1")((function(t){Array.from(t)})),"Array",{from:function(t){var e,a,r,f,d=o(t),g="function"==typeof this?this:Array,b=arguments.length,p=b>1?arguments[1]:void 0,m=void 0!==p,v=0,M=l(d);if(m&&(p=n(p,b>2?arguments[2]:void 0,2)),void 0==M||g==Array&&s(M))for(e=c(d.length),a=new g(e);e>v;v++)u(a,v,m?p(d[v],v):d[v]);else for(f=M.call(d),a=new g;!(r=f.next()).done;v++)u(a,v,m?i(f,p,[r.value,v],!0):r.value);return a.length=v,a}})},"54a1":function(t,e,a){a("6c1c"),a("1654"),t.exports=a("95d5")},"6b54":function(t,e,a){"use strict";a("3846");var n=a("cb7c"),r=a("0bfb"),o=a("9e1e"),i="toString",s=/./[i],c=function(t){a("2aba")(RegExp.prototype,i,t,!0)};a("79e5")((function(){return"/a/b"!=s.call({source:"a",flags:"b"})}))?c((function(){var t=n(this);return"/".concat(t.source,"/","flags"in t?t.flags:!o&&t instanceof RegExp?r.call(t):void 0)})):s.name!=i&&c((function(){return s.call(this)}))},"75fc":function(t,e,a){"use strict";var n=a("a745"),r=a.n(n);function o(t){if(r()(t)){for(var e=0,a=new Array(t.length);e<t.length;e++)a[e]=t[e];return a}}var i=a("774e"),s=a.n(i),c=a("c8bb"),u=a.n(c);function l(t){if(u()(Object(t))||"[object Arguments]"===Object.prototype.toString.call(t))return s()(t)}function f(){throw new TypeError("Invalid attempt to spread non-iterable instance")}function d(t){return o(t)||l(t)||f()}a.d(e,"a",(function(){return d}))},"774e":function(t,e,a){t.exports=a("d2d5")},"79b0":function(t,e,a){"use strict";var n=a("9e30"),r=a.n(n);r.a},9003:function(t,e,a){var n=a("6b4c");t.exports=Array.isArray||function(t){return"Array"==n(t)}},"95d5":function(t,e,a){var n=a("40c3"),r=a("5168")("iterator"),o=a("481b");t.exports=a("584a").isIterable=function(t){var e=Object(t);return void 0!==e[r]||"@@iterator"in e||o.hasOwnProperty(n(e))}},"9e30":function(t,e,a){},a448:function(t,e){t.exports="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0OCIgaGVpZ2h0PSI0MiI+CiAgPHBhdGggZmlsbD0iI0Q5RDlEOSIgZmlsbC1ydWxlPSJldmVub2RkIiBkPSJNNDggNHY1aC0yVjZIMnYzSDBWMkMwIC44OTU0MzA1Ljg5NTQzMSAwIDIgMGg0NGMxLjEwNDU2OSAwIDIgLjg5NTQzMDUgMiAydjJ6bS0yIDI2aC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDMwdjNIMHYtOGgydjNoNHYySDJ6bTQ0LTEyaC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDE4djNIMHYtOGgydjNoNHYySDJ6bTgtMmg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTEwIDI4aDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptMTIgMTRoLTR2LTJoNHYtM2gydjNjMCAxLjEwNDU2OTUtLjg5NTQzMSAyLTIgMnpNMiA0MGg0djJIMmMtMS4xMDQ1NjkgMC0yLS44OTU0MzA1LTItMnYtM2gydjN6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTIgMnYyaDQ0VjJIMnoiLz4KPC9zdmc+Cg=="},a745:function(t,e,a){t.exports=a("f410")},be10:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return t.metrics?a("div",{staticClass:"info-grid"},t._l(t.metrics,(function(e,n){return null!==e.value?a("div",{key:n,staticClass:"metric",class:e.status,attrs:{"data-testid":e.metric}},[a("span",{staticClass:"metric-title"},[t._v(t._s(e.metric))]),a("span",{staticClass:"metric-value",class:{"has-error":n===t.hasError[n]}},[t._v(t._s(t._f("formatError")(t._f("formatValue")(e.value))))])]):t._e()})),0):t._e()},r=[],o=(a("456d"),a("ac6a"),a("6b54"),{name:"MetricsGrid",filters:{formatValue:function(t){return t?t.toLocaleString("en").toString():0},formatError:function(t){return"--"===t?"error calculating":t}},props:{metrics:{type:Array,required:!0,default:function(){}}},computed:{hasError:function(){var t=this,e={};return Object.keys(this.metrics).forEach((function(a){"--"===t.metrics[a].value&&(e[a]=a)})),e}}}),i=o,s=(a("79b0"),a("2877")),c=Object(s["a"])(i,n,r,!1,null,null,null);e["a"]=c.exports},c4a3:function(t,e,a){"use strict";var n=a("c74b"),r=a.n(n);r.a},c74b:function(t,e,a){},c8bb:function(t,e,a){t.exports=a("54a1")},d2d5:function(t,e,a){a("1654"),a("549b"),t.exports=a("584a").Array.from},e234:function(t,e,a){"use strict";var n=a("2055"),r=a.n(n);r.a},f410:function(t,e,a){a("1af6"),t.exports=a("584a").Array.isArray}}]);
//# sourceMappingURL=chunk-4bd71e6e.fecb8836.js.map