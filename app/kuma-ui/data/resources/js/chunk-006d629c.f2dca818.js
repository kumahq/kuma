(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-006d629c"],{"0bfb":function(t,a,e){"use strict";var n=e("cb7c");t.exports=function(){var t=n(this),a="";return t.global&&(a+="g"),t.ignoreCase&&(a+="i"),t.multiline&&(a+="m"),t.unicode&&(a+="u"),t.sticky&&(a+="y"),a}},"1af6":function(t,a,e){var n=e("63b6");n(n.S,"Array",{isArray:e("9003")})},"20fd":function(t,a,e){"use strict";var n=e("d9f6"),r=e("aebd");t.exports=function(t,a,e){a in t?n.f(t,a,r(0,e)):t[a]=e}},2699:function(t,a,e){},2778:function(t,a,e){"use strict";var n=function(){var t=this,a=t.$createElement,n=t._self._c||a;return n("div",{staticClass:"data-overview"},[t.isReady?n("div",{staticClass:"data-overview-content"},[!t.isLoading&&t.displayMetrics&&t.metricsData?n("MetricGrid",{attrs:{metrics:t.metricsData}}):t.isLoading&&t.displayMetrics?n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[t._v("\n        "+t._s(t.emptyState.title)+"\n      ")]),t.showCta?n("template",{slot:"message"},[t.ctaAction&&t.ctaAction.length?n("router-link",{attrs:{to:t.ctaAction}},[t._v("\n          "+t._s(t.emptyState.ctaText)+"\n        ")]):t._e(),t._v("\n        "+t._s(t.emptyState.message)+"\n      ")],1):t._e()],2):t._e(),t.displayDataTable&&!1===t.tableDataIsEmpty&&t.tableData?n("KTable",{attrs:{options:t.tableData},scopedSlots:t._u([{key:"actions",fn:function(a){var e=a.row;return[n("router-link",{attrs:{to:{name:t.tableActionsRouteName,params:{mesh:"Mesh"===e.type?e.name:e.mesh,dataplane:"Dataplane"===e.type?e.name:null}}}},[t._t("tableDataActionsLinkText")],2)]}}],null,!0)}):t._e(),!0===t.tableDataIsEmpty?n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[n("div",{staticClass:"card-icon mb-3"},[n("img",{attrs:{src:e("a448")}})]),t._v("\n        No Items Found\n      ")])],2):t._e(),t.$slots.content?n("div",{staticClass:"data-overview-content mt-6"},[t._t("content")],2):t._e()],1):n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"48"}})],1),t._v("\n      Data Loading...\n    ")])],2)],1)},r=[],i=e("be10"),s={name:"DataOverview",components:{MetricGrid:i["a"]},props:{displayMetrics:{type:Boolean,default:!1},metricsData:{type:Array,default:null},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},emptyState:{type:Object,default:null},ctaAction:{type:Object,default:function(){}},showCta:{type:Boolean,default:!0},displayDataTable:{type:Boolean,default:!1},tableData:{type:Object,default:null},tableDataIsEmpty:{type:Boolean,default:!1},tableDataActionsLink:{type:String,default:null},tableActionsRouteName:{type:String,default:null}},computed:{isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading}}},o=s,c=(e("9947"),e("2877")),l=Object(c["a"])(o,n,r,!1,null,null,null);a["a"]=l.exports},3846:function(t,a,e){e("9e1e")&&"g"!=/./g.flags&&e("86cc").f(RegExp.prototype,"flags",{configurable:!0,get:e("0bfb")})},"42f1":function(t,a,e){"use strict";e.r(a);var n=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("div",{staticClass:"traffic-permissions"},[e("DataOverview",{attrs:{"display-data-table":!0,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty}})],1)},r=[],i=e("75fc"),s=e("2778"),o={name:"TrafficPermissions",metaInfo:{title:"Traffic Permissions"},components:{DataOverview:s["a"]},data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,tableDataIsEmpty:!1,tableData:{headers:[{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"}],data:[]}}},watch:{$route:function(t,a){this.bootstrap()}},beforeMount:function(){this.bootstrap()},methods:{bootstrap:function(){var t=this;this.isLoading=!0,this.isEmpty=!1;var a=this.$route.params.mesh,e=function(){return t.$api.getTrafficPermissions(a).then((function(a){var e=a.items;e&&e.length?(t.tableData.data=Object(i["a"])(e),t.tableDataIsEmpty=!1):(t.tableData.data=[],t.tableDataIsEmpty=!0)})).catch((function(a){t.tableDataIsEmpty=!0,t.isEmpty=!0,console.error(a)}))};e()}}},c=o,l=e("2877"),u=Object(l["a"])(c,n,r,!1,null,null,null);a["default"]=u.exports},"549b":function(t,a,e){"use strict";var n=e("d864"),r=e("63b6"),i=e("241e"),s=e("b0dc"),o=e("3702"),c=e("b447"),l=e("20fd"),u=e("7cd6");r(r.S+r.F*!e("4ee1")((function(t){Array.from(t)})),"Array",{from:function(t){var a,e,r,f,d=i(t),p="function"==typeof this?this:Array,m=arguments.length,y=m>1?arguments[1]:void 0,b=void 0!==y,g=0,v=u(d);if(b&&(y=n(y,m>2?arguments[2]:void 0,2)),void 0==v||p==Array&&o(v))for(a=c(d.length),e=new p(a);a>g;g++)l(e,g,b?y(d[g],g):d[g]);else for(f=v.call(d),e=new p;!(r=f.next()).done;g++)l(e,g,b?s(f,y,[r.value,g],!0):r.value);return e.length=g,e}})},"54a1":function(t,a,e){e("6c1c"),e("1654"),t.exports=e("95d5")},"6b54":function(t,a,e){"use strict";e("3846");var n=e("cb7c"),r=e("0bfb"),i=e("9e1e"),s="toString",o=/./[s],c=function(t){e("2aba")(RegExp.prototype,s,t,!0)};e("79e5")((function(){return"/a/b"!=o.call({source:"a",flags:"b"})}))?c((function(){var t=n(this);return"/".concat(t.source,"/","flags"in t?t.flags:!i&&t instanceof RegExp?r.call(t):void 0)})):o.name!=s&&c((function(){return o.call(this)}))},"75fc":function(t,a,e){"use strict";var n=e("a745"),r=e.n(n);function i(t){if(r()(t)){for(var a=0,e=new Array(t.length);a<t.length;a++)e[a]=t[a];return e}}var s=e("774e"),o=e.n(s),c=e("c8bb"),l=e.n(c);function u(t){if(l()(Object(t))||"[object Arguments]"===Object.prototype.toString.call(t))return o()(t)}function f(){throw new TypeError("Invalid attempt to spread non-iterable instance")}function d(t){return i(t)||u(t)||f()}e.d(a,"a",(function(){return d}))},"774e":function(t,a,e){t.exports=e("d2d5")},"79b0":function(t,a,e){"use strict";var n=e("9e30"),r=e.n(n);r.a},9003:function(t,a,e){var n=e("6b4c");t.exports=Array.isArray||function(t){return"Array"==n(t)}},"95d5":function(t,a,e){var n=e("40c3"),r=e("5168")("iterator"),i=e("481b");t.exports=e("584a").isIterable=function(t){var a=Object(t);return void 0!==a[r]||"@@iterator"in a||i.hasOwnProperty(n(a))}},9947:function(t,a,e){"use strict";var n=e("2699"),r=e.n(n);r.a},"9e30":function(t,a,e){},a448:function(t,a){t.exports="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0OCIgaGVpZ2h0PSI0MiI+CiAgPHBhdGggZmlsbD0iI0Q5RDlEOSIgZmlsbC1ydWxlPSJldmVub2RkIiBkPSJNNDggNHY1aC0yVjZIMnYzSDBWMkMwIC44OTU0MzA1Ljg5NTQzMSAwIDIgMGg0NGMxLjEwNDU2OSAwIDIgLjg5NTQzMDUgMiAydjJ6bS0yIDI2aC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDMwdjNIMHYtOGgydjNoNHYySDJ6bTQ0LTEyaC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDE4djNIMHYtOGgydjNoNHYySDJ6bTgtMmg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTEwIDI4aDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptMTIgMTRoLTR2LTJoNHYtM2gydjNjMCAxLjEwNDU2OTUtLjg5NTQzMSAyLTIgMnpNMiA0MGg0djJIMmMtMS4xMDQ1NjkgMC0yLS44OTU0MzA1LTItMnYtM2gydjN6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTIgMnYyaDQ0VjJIMnoiLz4KPC9zdmc+Cg=="},a745:function(t,a,e){t.exports=e("f410")},be10:function(t,a,e){"use strict";var n=function(){var t=this,a=t.$createElement,e=t._self._c||a;return t.metrics?e("div",{staticClass:"info-grid"},t._l(t.metrics,(function(a,n){return null!==a.value?e("div",{key:n,staticClass:"metric",class:a.status,attrs:{"data-testid":a.metric}},[e("span",{staticClass:"metric-title"},[t._v(t._s(a.metric))]),e("span",{staticClass:"metric-value",class:{"has-error":n===t.hasError[n]}},[t._v(t._s(t._f("formatError")(t._f("formatValue")(a.value))))])]):t._e()})),0):t._e()},r=[],i=(e("456d"),e("ac6a"),e("6b54"),{name:"MetricsGrid",filters:{formatValue:function(t){return t?t.toLocaleString("en").toString():0},formatError:function(t){return"--"===t?"error calculating":t}},props:{metrics:{type:Array,required:!0,default:function(){}}},computed:{hasError:function(){var t=this,a={};return Object.keys(this.metrics).forEach((function(e){"--"===t.metrics[e].value&&(a[e]=e)})),a}}}),s=i,o=(e("79b0"),e("2877")),c=Object(o["a"])(s,n,r,!1,null,null,null);a["a"]=c.exports},c8bb:function(t,a,e){t.exports=e("54a1")},d2d5:function(t,a,e){e("1654"),e("549b"),t.exports=e("584a").Array.from},f410:function(t,a,e){e("1af6"),t.exports=e("584a").Array.isArray}}]);
//# sourceMappingURL=chunk-006d629c.f2dca818.js.map