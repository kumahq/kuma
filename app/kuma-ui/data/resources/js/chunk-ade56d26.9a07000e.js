(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-ade56d26"],{"0eaf":function(t,a,e){"use strict";e.r(a);var n=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("div",{staticClass:"dataplanes"},[e("DataOverview",{attrs:{"empty-state":t.empty_state,"display-data-table":!0,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty}})],1)},s=[],i=(e("a481"),e("7f7f"),e("ac6a"),e("2778")),r={name:"Dataplanes",components:{DataOverview:i["a"]},data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no items present."},tableData:{headers:[{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Tags",key:"tags"},{label:"Status",key:"status"},{label:"Last Connected",key:"lastConnected"},{label:"Last Updated",key:"lastUpdated"},{label:"Total Updates",key:"totalUpdates"}],data:[]}}},watch:{$route:function(t,a){this.bootstrap()}},beforeMount:function(){this.bootstrap()},methods:{bootstrap:function(){var t=this;this.isLoading=!0,this.isEmpty=!1;var a=this.$route.params.mesh,e=function(){return t.$api.getAllDataplanesFromMesh(a).then((function(e){var n=e.items,s=[];n.forEach((function(e){t.$api.getDataplaneOverviews(a,e.name).then((function(t){var a,e,n,i=new Date,r=[],o="Offline",l=[],c=[];if(t.dataplane.networking.inbound&&t.dataplane.networking.inbound.length)for(var u=0;u<t.dataplane.networking.inbound.length;u++){var d=t.dataplane.networking.inbound[u].tags;n=JSON.stringify(d).replace(/[{}]/g,"").replace(/"/g,"").replace(/,/g,", ")}if(t.dataplaneInsight.subscriptions&&t.dataplaneInsight.subscriptions.length){t.dataplaneInsight.subscriptions.forEach((function(t){r.push(t.status.total.responsesSent),l.push(t.connectTime),c.push(t.status.lastUpdateTime),o=t.connectTime&&t.connectTime.length&&!t.disconnectTime?"Online":"Offline"})),r=r.reduce((function(t,a){return t+a}));var g=new Date(l.reduce((function(t,a){return t.MeasureDate>a.MeasureDate?t:a}))),p=new Date(c.reduce((function(t,a){return t.MeasureDate>a.MeasureDate?t:a})));a="".concat(Math.abs(i.getHours()-g.getHours()),"h ").concat(Math.abs(i.getMinutes()-g.getMinutes()),"m ").concat(Math.abs(i.getSeconds()-g.getSeconds()),"s"),e="".concat(Math.abs(i.getHours()-p.getHours()),"h ").concat(Math.abs(i.getMinutes()-p.getMinutes()),"m ").concat(Math.abs(i.getSeconds()-p.getSeconds()),"s")}else a=e=r="n/a";s.push({name:t.name,mesh:t.mesh,tags:n,status:o,lastConnected:a,lastUpdated:e,totalUpdates:r})})).catch((function(t){console.error(t)}))})),n&&n.length?(t.tableData.data=s,t.tableDataIsEmpty=!1):(t.tableData.data=[],t.tableDataIsEmpty=!0)})).catch((function(a){t.tableDataIsEmpty=!0,t.isEmpty=!0,console.error(a)}))};e()}}},o=r,l=e("2877"),c=Object(l["a"])(o,n,s,!1,null,null,null);a["default"]=c.exports},2699:function(t,a,e){},2778:function(t,a,e){"use strict";var n=function(){var t=this,a=t.$createElement,n=t._self._c||a;return n("div",{staticClass:"data-overview"},[t.isReady?n("div",{staticClass:"data-overview-content"},[!t.isLoading&&t.displayMetrics&&t.metricsData?n("MetricGrid",{attrs:{metrics:t.metricsData}}):t.isLoading&&t.displayMetrics?n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[t._v("\n        "+t._s(t.emptyState.title)+"\n      ")]),t.showCta?n("template",{slot:"message"},[t.ctaAction&&t.ctaAction.length?n("router-link",{attrs:{to:t.ctaAction}},[t._v("\n          "+t._s(t.emptyState.ctaText)+"\n        ")]):t._e(),t._v("\n        "+t._s(t.emptyState.message)+"\n      ")],1):t._e()],2):t._e(),t.displayDataTable&&!1===t.tableDataIsEmpty&&t.tableData?n("KTable",{attrs:{options:t.tableData},scopedSlots:t._u([{key:"actions",fn:function(a){var e=a.row;return[n("router-link",{attrs:{to:{name:t.tableActionsRouteName,params:{mesh:"Mesh"===e.type?e.name:e.mesh,dataplane:"Dataplane"===e.type?e.name:null}}}},[t._t("tableDataActionsLinkText")],2)]}}],null,!0)}):t._e(),!0===t.tableDataIsEmpty?n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[n("div",{staticClass:"card-icon mb-3"},[n("img",{attrs:{src:e("a448")}})]),t._v("\n        No Items Found\n      ")])],2):t._e(),t.$slots.content?n("div",{staticClass:"data-overview-content mt-6"},[t._t("content")],2):t._e()],1):n("KEmptyState",{attrs:{"cta-is-hidden":""}},[n("template",{slot:"title"},[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"48"}})],1),t._v("\n      Data Loading...\n    ")])],2)],1)},s=[],i=e("be10"),r={name:"DataOverview",components:{MetricGrid:i["a"]},props:{displayMetrics:{type:Boolean,default:!1},metricsData:{type:Array,default:null},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},emptyState:{type:Object,default:null},ctaAction:{type:Object,default:function(){}},showCta:{type:Boolean,default:!0},displayDataTable:{type:Boolean,default:!1},tableData:{type:Object,default:null},tableDataIsEmpty:{type:Boolean,default:!1},tableDataActionsLink:{type:String,default:null},tableActionsRouteName:{type:String,default:null}},computed:{isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading}}},o=r,l=(e("9947"),e("2877")),c=Object(l["a"])(o,n,s,!1,null,null,null);a["a"]=c.exports},3846:function(t,a,e){e("9e1e")&&"g"!=/./g.flags&&e("86cc").f(RegExp.prototype,"flags",{configurable:!0,get:e("0bfb")})},"6b54":function(t,a,e){"use strict";e("3846");var n=e("cb7c"),s=e("0bfb"),i=e("9e1e"),r="toString",o=/./[r],l=function(t){e("2aba")(RegExp.prototype,r,t,!0)};e("79e5")((function(){return"/a/b"!=o.call({source:"a",flags:"b"})}))?l((function(){var t=n(this);return"/".concat(t.source,"/","flags"in t?t.flags:!i&&t instanceof RegExp?s.call(t):void 0)})):o.name!=r&&l((function(){return o.call(this)}))},"79b0":function(t,a,e){"use strict";var n=e("9e30"),s=e.n(n);s.a},9947:function(t,a,e){"use strict";var n=e("2699"),s=e.n(n);s.a},"9e30":function(t,a,e){},a448:function(t,a){t.exports="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0OCIgaGVpZ2h0PSI0MiI+CiAgPHBhdGggZmlsbD0iI0Q5RDlEOSIgZmlsbC1ydWxlPSJldmVub2RkIiBkPSJNNDggNHY1aC0yVjZIMnYzSDBWMkMwIC44OTU0MzA1Ljg5NTQzMSAwIDIgMGg0NGMxLjEwNDU2OSAwIDIgLjg5NTQzMDUgMiAydjJ6bS0yIDI2aC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDMwdjNIMHYtOGgydjNoNHYySDJ6bTQ0LTEyaC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDE4djNIMHYtOGgydjNoNHYySDJ6bTgtMmg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTEwIDI4aDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptMTIgMTRoLTR2LTJoNHYtM2gydjNjMCAxLjEwNDU2OTUtLjg5NTQzMSAyLTIgMnpNMiA0MGg0djJIMmMtMS4xMDQ1NjkgMC0yLS44OTU0MzA1LTItMnYtM2gydjN6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTIgMnYyaDQ0VjJIMnoiLz4KPC9zdmc+Cg=="},a481:function(t,a,e){"use strict";var n=e("cb7c"),s=e("4bf8"),i=e("9def"),r=e("4588"),o=e("0390"),l=e("5f1b"),c=Math.max,u=Math.min,d=Math.floor,g=/\$([$&`']|\d\d?|<[^>]*>)/g,p=/\$([$&`']|\d\d?)/g,f=function(t){return void 0===t?t:String(t)};e("214f")("replace",2,(function(t,a,e,m){return[function(n,s){var i=t(this),r=void 0==n?void 0:n[a];return void 0!==r?r.call(n,i,s):e.call(String(i),n,s)},function(t,a){var s=m(e,t,this,a);if(s.done)return s.value;var d=n(t),g=String(this),p="function"===typeof a;p||(a=String(a));var h=d.global;if(h){var y=d.unicode;d.lastIndex=0}var v=[];while(1){var M=l(d,g);if(null===M)break;if(v.push(M),!h)break;var D=String(M[0]);""===D&&(d.lastIndex=o(g,i(d.lastIndex),y))}for(var T="",I=0,L=0;L<v.length;L++){M=v[L];for(var S=String(M[0]),E=c(u(r(M.index),g.length),0),w=[],_=1;_<M.length;_++)w.push(f(M[_]));var j=M.groups;if(p){var N=[S].concat(w,E,g);void 0!==j&&N.push(j);var C=String(a.apply(void 0,N))}else C=b(S,g,E,w,j,a);E>=I&&(T+=g.slice(I,E)+C,I=E+S.length)}return T+g.slice(I)}];function b(t,a,n,i,r,o){var l=n+t.length,c=i.length,u=p;return void 0!==r&&(r=s(r),u=g),e.call(o,u,(function(e,s){var o;switch(s.charAt(0)){case"$":return"$";case"&":return t;case"`":return a.slice(0,n);case"'":return a.slice(l);case"<":o=r[s.slice(1,-1)];break;default:var u=+s;if(0===u)return e;if(u>c){var g=d(u/10);return 0===g?e:g<=c?void 0===i[g-1]?s.charAt(1):i[g-1]+s.charAt(1):e}o=i[u-1]}return void 0===o?"":o}))}}))},be10:function(t,a,e){"use strict";var n=function(){var t=this,a=t.$createElement,e=t._self._c||a;return t.metrics?e("div",{staticClass:"info-grid"},t._l(t.metrics,(function(a,n){return null!==a.value?e("div",{key:n,staticClass:"metric",class:a.status,attrs:{"data-testid":a.metric}},[e("span",{staticClass:"metric-title"},[t._v(t._s(a.metric))]),e("span",{staticClass:"metric-value",class:{"has-error":n===t.hasError[n]}},[t._v(t._s(t._f("formatError")(t._f("formatValue")(a.value))))])]):t._e()})),0):t._e()},s=[],i=(e("456d"),e("ac6a"),e("6b54"),{name:"MetricsGrid",filters:{formatValue:function(t){return t?t.toLocaleString("en").toString():0},formatError:function(t){return"--"===t?"error calculating":t}},props:{metrics:{type:Array,required:!0,default:function(){}}},computed:{hasError:function(){var t=this,a={};return Object.keys(this.metrics).forEach((function(e){"--"===t.metrics[e].value&&(a[e]=e)})),a}}}),r=i,o=(e("79b0"),e("2877")),l=Object(o["a"])(r,n,s,!1,null,null,null);a["a"]=l.exports}}]);
//# sourceMappingURL=chunk-ade56d26.9a07000e.js.map