(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["dataplanes-ingress"],{"4a34":function(t,e,n){"use strict";n.d(e,"a",(function(){return r})),n.d(e,"d",(function(){return s})),n.d(e,"b",(function(){return o})),n.d(e,"c",(function(){return l}));n("99af"),n("4de4"),n("5db7"),n("a630"),n("d81d"),n("45fc"),n("73d9"),n("b0c0"),n("4fad"),n("d3b7"),n("ac1f"),n("6062"),n("3ca3"),n("1276"),n("ddb0");var a=n("5530"),i=n("3835");function r(t){var e=[],n=t.networking.inbound||null;n&&(e=n.flatMap((function(t){return Object.entries(t.tags)})).map((function(t){var e=Object(i["a"])(t,2),n=e[0],a=e[1];return n+"="+a})));var a=t.networking.gateway||null;return a&&(e=Object.entries(a.tags).map((function(t){var e=Object(i["a"])(t,2),n=e[0],a=e[1];return n+"="+a}))),e=Array.from(new Set(e)),e.map((function(t){return t.split("=")})).map((function(t){var e=Object(i["a"])(t,2),n=e[0],a=e[1];return{label:n,value:a}}))}function s(t){var e=arguments.length>1&&void 0!==arguments[1]?arguments[1]:{},n=t.networking.inbound?t.networking.inbound:[{health:{ready:!0}}],a=n.filter((function(t){return t.health&&!t.health.ready})).map((function(t){return"Inbound on port ".concat(t.port," is not ready (kuma.io/service: ").concat(t.tags["kuma.io/service"],")")})),i=e.subscriptions?e.subscriptions:[],r=i.some((function(t){return t.connectTime&&t.connectTime.length&&!t.disconnectTime})),s=function(){var t=a.length===n.length,e=0===a.length;return!r||t?"Offline":e?"Online":"Partially degraded"};return{status:s(),reason:a}}function o(t){var e=t.name,n=t.mesh,i=t.type;return Object(a["a"])({name:e,mesh:n,type:i},t.dataplane)}function l(t){var e=t.name,n=t.mesh,i=t.type;return Object(a["a"])({name:e,mesh:n,type:i},t.dataplaneInsight)}},"5db7":function(t,e,n){"use strict";var a=n("23e7"),i=n("a2bf"),r=n("7b0b"),s=n("50c4"),o=n("1c0b"),l=n("65f0");a({target:"Array",proto:!0},{flatMap:function(t){var e,n=r(this),a=s(n.length);return o(t),e=l(n,0),e.length=i(e,n,n,a,0,1,t,arguments.length>1?arguments[1]:void 0),e}})},6062:function(t,e,n){"use strict";var a=n("6d61"),i=n("6566");t.exports=a("Set",(function(t){return function(){return t(this,arguments.length?arguments[0]:void 0)}}),i)},6566:function(t,e,n){"use strict";var a=n("9bf2").f,i=n("7c73"),r=n("e2cc"),s=n("0366"),o=n("19aa"),l=n("2266"),u=n("7dd0"),c=n("2626"),d=n("83ab"),p=n("f183").fastKey,f=n("69f3"),v=f.set,h=f.getterFor;t.exports={getConstructor:function(t,e,n,u){var c=t((function(t,a){o(t,c,e),v(t,{type:e,index:i(null),first:void 0,last:void 0,size:0}),d||(t.size=0),void 0!=a&&l(a,t[u],{that:t,AS_ENTRIES:n})})),f=h(e),b=function(t,e,n){var a,i,r=f(t),s=y(t,e);return s?s.value=n:(r.last=s={index:i=p(e,!0),key:e,value:n,previous:a=r.last,next:void 0,removed:!1},r.first||(r.first=s),a&&(a.next=s),d?r.size++:t.size++,"F"!==i&&(r.index[i]=s)),t},y=function(t,e){var n,a=f(t),i=p(e);if("F"!==i)return a.index[i];for(n=a.first;n;n=n.next)if(n.key==e)return n};return r(c.prototype,{clear:function(){var t=this,e=f(t),n=e.index,a=e.first;while(a)a.removed=!0,a.previous&&(a.previous=a.previous.next=void 0),delete n[a.index],a=a.next;e.first=e.last=void 0,d?e.size=0:t.size=0},delete:function(t){var e=this,n=f(e),a=y(e,t);if(a){var i=a.next,r=a.previous;delete n.index[a.index],a.removed=!0,r&&(r.next=i),i&&(i.previous=r),n.first==a&&(n.first=i),n.last==a&&(n.last=r),d?n.size--:e.size--}return!!a},forEach:function(t){var e,n=f(this),a=s(t,arguments.length>1?arguments[1]:void 0,3);while(e=e?e.next:n.first){a(e.value,e.key,this);while(e&&e.removed)e=e.previous}},has:function(t){return!!y(this,t)}}),r(c.prototype,n?{get:function(t){var e=y(this,t);return e&&e.value},set:function(t,e){return b(this,0===t?0:t,e)}}:{add:function(t){return b(this,t=0===t?0:t,t)}}),d&&a(c.prototype,"size",{get:function(){return f(this).size}}),c},setStrong:function(t,e,n){var a=e+" Iterator",i=h(e),r=h(a);u(t,e,(function(t,e){v(this,{type:a,target:t,state:i(t),kind:e,last:void 0})}),(function(){var t=r(this),e=t.kind,n=t.last;while(n&&n.removed)n=n.previous;return t.target&&(t.last=n=n?n.next:t.state.first)?"keys"==e?{value:n.key,done:!1}:"values"==e?{value:n.value,done:!1}:{value:[n.key,n.value],done:!1}:(t.target=void 0,{value:void 0,done:!0})}),n?"entries":"values",!n,!0),c(e)}}},6663:function(t,e,n){"use strict";var a=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"entity-url-control"},[t.shouldDisplay?n("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function(e){var a=e.copyToClipboard;return[n("KPop",{attrs:{placement:"bottom"}},[n("KButton",{attrs:{appearance:"secondary",size:"small"},on:{click:function(){a(t.url)}}},[n("KIcon",{attrs:{slot:"icon",icon:"externalLink"},slot:"icon"}),t._v(" "+t._s(t.copyButtonText)+" ")],1),n("div",{attrs:{slot:"content"},slot:"content"},[n("p",[t._v(t._s(t.confirmationText))])])],1)]}}],null,!1,1603401634)}):t._e()],1)},i=[],r={name:"EntityURLControl",props:{url:{type:String,required:!0},copyButtonText:{type:String,default:"Copy URL"},confirmationText:{type:String,default:"URL copied to clipboard!"}},computed:{shouldDisplay:function(){var t=this.$route.params.mesh||null;return!(!t||"all"===t)}}},s=r,o=n("2877"),l=Object(o["a"])(s,a,i,!1,null,null,null);e["a"]=l.exports},"6d61":function(t,e,n){"use strict";var a=n("23e7"),i=n("da84"),r=n("94ca"),s=n("6eeb"),o=n("f183"),l=n("2266"),u=n("19aa"),c=n("861d"),d=n("d039"),p=n("1c7e"),f=n("d44e"),v=n("7156");t.exports=function(t,e,n){var h=-1!==t.indexOf("Map"),b=-1!==t.indexOf("Weak"),y=h?"set":"add",m=i[t],g=m&&m.prototype,w=m,E={},x=function(t){var e=g[t];s(g,t,"add"==t?function(t){return e.call(this,0===t?0:t),this}:"delete"==t?function(t){return!(b&&!c(t))&&e.call(this,0===t?0:t)}:"get"==t?function(t){return b&&!c(t)?void 0:e.call(this,0===t?0:t)}:"has"==t?function(t){return!(b&&!c(t))&&e.call(this,0===t?0:t)}:function(t,n){return e.call(this,0===t?0:t,n),this})};if(r(t,"function"!=typeof m||!(b||g.forEach&&!d((function(){(new m).entries().next()})))))w=n.getConstructor(e,t,h,y),o.REQUIRED=!0;else if(r(t,!0)){var O=new w,D=O[y](b?{}:-0,1)!=O,k=d((function(){O.has(1)})),_=p((function(t){new m(t)})),I=!b&&d((function(){var t=new m,e=5;while(e--)t[y](e,e);return!t.has(-0)}));_||(w=e((function(e,n){u(e,w,t);var a=v(new m,e,w);return void 0!=n&&l(n,a[y],{that:a,AS_ENTRIES:h}),a})),w.prototype=g,g.constructor=w),(k||I)&&(x("delete"),x("has"),h&&x("get")),(I||D)&&x(y),b&&g.clear&&delete g.clear}return E[t]=w,a({global:!0,forced:w!=m},E),f(w,t),b||n.setStrong(w,t,h),w}},"73d9":function(t,e,n){var a=n("44d2");a("flatMap")},"748e":function(t,e,n){"use strict";n.r(e);var a=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"ingress-dataplanes"},[n("FrameSkeleton",[n("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"display-data-table":!0,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,"table-data-function-text":"View","table-data-row":"name"},on:{tableAction:t.tableAction,reloadData:t.loadData}},[n("template",{slot:"additionalControls"},[n("KButton",{staticClass:"add-dp-button",attrs:{appearance:"primary",size:"small",to:t.dataplaneWizardRoute}},[n("span",{staticClass:"custom-control-icon"},[t._v(" + ")]),t._v(" Create Dataplane Proxy ")]),this.$route.query.ns?n("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:"ingress-dataplanes"}}},[n("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()],1),n("template",{slot:"pagination"},[n("Pagination",{attrs:{"has-previous":t.previous.length>0,"has-next":t.hasNext},on:{next:t.goToNextPage,previous:t.goToPreviousPage}})],1)],2),!1===t.isEmpty?n("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"}},[n("template",{slot:"tabHeader"},[n("div",[n("h3",[t._v(t._s(t.tabGroupTitle))])]),n("div",[n("EntityURLControl",{attrs:{url:t.shareUrl}})],1)]),n("template",{slot:"overview"},[n("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[n("div",[n("ul",t._l(t.entity.basicData,(function(e,a){return n("li",{key:a},[n("div","status"===a?[n("h4",[t._v(t._s(a))]),n("div",{staticClass:"entity-status",class:{"is-offline":"offline"===e.status.toString().toLowerCase()||!1===e.status,"is-degraded":"partially degraded"===e.status.toString().toLowerCase()||!1===e.status}},[n("span",{staticClass:"entity-status__label"},[t._v(t._s(e.status))])]),n("div",{staticClass:"reason-list"},[n("ul",t._l(e.reason,(function(e){return n("li",{key:e},[n("span",{staticClass:"entity-status__dot"}),t._v(" "+t._s(e)+" ")])})),0)])]:[n("h4",[t._v(t._s(a))]),t._v(" "+t._s(e)+" ")])])})),0)]),n("div",[n("h4",[t._v("Tags")]),n("p",t._l(t.entity.tags,(function(e,a){return n("span",{key:a,staticClass:"tag-cols"},[n("span",[t._v(" "+t._s(e.label)+": ")]),n("span",[t._v(" "+t._s(e.value)+" ")])])})),0)])])],1),n("template",{slot:"yaml"},[n("YamlView",{attrs:{title:t.entityOverviewTitle,"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})],1)],2):t._e()],1)],1)},i=[],r=(n("99af"),n("4160"),n("13d5"),n("b0c0"),n("d3b7"),n("159b"),n("96cf"),n("1da1")),s=n("5530"),o=n("2f62"),l=n("d7c2"),u=n("4a34"),c=n("6663"),d=n("8218"),p=n("1d10"),f=n("1799"),v=n("2778"),h=n("251b"),b=n("ff9d"),y=n("0ada"),m={name:"IngressDataplanes",metaInfo:{title:"Ingress data plane proxies"},components:{EntityURLControl:c["a"],FrameSkeleton:p["a"],Pagination:f["a"],DataOverview:v["a"],Tabs:h["a"],YamlView:b["a"],LabelList:y["a"]},mixins:[d["a"]],data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Ingress data plane proxies present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Type",key:"type"},{label:"Tags",key:"tags"},{label:"Public address",key:"publicAddress"},{label:"Public port",key:"publicPort"},{label:"Last Connected",key:"lastConnected"},{label:"Last Updated",key:"lastUpdated"},{label:"Total Updates",key:"totalUpdates"},{label:"Kuma DP version",key:"dpVersion"},{label:"Envoy version",key:"envoyVersion"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#yaml",title:"YAML"}],entity:[],rawEntity:null,firstEntity:null,pageSize:this.$pageSize,pageOffset:null,next:null,hasNext:!1,previous:[],tabGroupTitle:null,entityNamespace:null,entityOverviewTitle:null,showmTLSTab:!1}},computed:Object(s["a"])(Object(s["a"])({},Object(o["c"])({environment:"getEnvironment",queryNamespace:"getItemQueryNamespace"})),{},{dataplaneWizardRoute:function(){return"universal"===this.environment?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}},version:function(){var t=this.$store.getters.getVersion;return null!==t?t:"latest"},shareUrl:function(){var t=this,e="".concat(window.location.origin,"#"),n=this.entity,a=function(){return n.basicData?t.$route.query.ns?t.$route.fullPath:"".concat(e).concat(t.$route.fullPath,"?ns=").concat(n.basicData.name):null};return a()}}),watch:{$route:function(t,e){this.loadData()}},beforeMount:function(){this.loadData()},methods:{init:function(){this.loadData()},goToPreviousPage:function(){this.pageOffset=this.previous.pop(),this.next=null,this.loadData()},goToNextPage:function(){this.previous.push(this.pageOffset),this.pageOffset=this.next,this.next=null,this.loadData()},tableAction:function(t){var e=t;this.$store.dispatch("updateSelectedTab",this.tabs[0].hash),this.$store.dispatch("updateSelectedTableRow",e.name),this.getEntity(e)},loadData:function(){var t=this;this.isLoading=!0;var e=this.$route.params.mesh||null,n=this.$route.query.ns||null,a={size:this.pageSize,offset:this.pageOffset,ingress:!0},i=function(){return"all"===e?t.$api.getAllDataplaneOverviews(a):n&&n.length&&"all"!==e?t.$api.getDataplaneOverviewFromMesh(e,n):t.$api.getAllDataplaneOverviewsFromMesh(e,a)},r=function(e,n,a){t.$api.getDataplaneOverviewFromMesh(e,n).then((function(e){var n,i,r="n/a",s=[],o=[],c=[],d="",p="",f=[],v=[],h=(e.dataplane.networking.ingress,"Ingress");s=Object(u["a"])(e.dataplane);var b=Object(u["d"])(e.dataplane,e.dataplaneInsight),y=b.status;if(e.dataplaneInsight.subscriptions&&e.dataplaneInsight.subscriptions.length){e.dataplaneInsight.subscriptions.forEach((function(t){var e=t.status.total.responsesSent||0,n=t.status.total.responsesRejected||0,a=t.connectTime||r,i=t.status.lastUpdateTime||r;o.push(parseInt(e)),c.push(parseInt(n)),f.push(a),v.push(i),t.version&&t.version.kumaDp&&(d=t.version.kumaDp.version,p=t.version.envoy.version)})),o=o.reduce((function(t,e){return t+e})),c=c.reduce((function(t,e){return t+e}));var m=f.reduce((function(t,e){return t&&e?t.MeasureDate>e.MeasureDate?t:e:null})),g=v.reduce((function(t,e){return t&&e?t.MeasureDate>e.MeasureDate?t:e:null})),w=new Date(m),E=new Date(g);n=m&&!isNaN(w)?Object(l["g"])(w):"never",i=g&&!isNaN(E)?Object(l["g"])(E):"never"}else n="never",i="never",o=0,c=0,d="-",p="-";return a.push({name:e.name,mesh:e.mesh,tags:s,status:y,lastConnected:n,lastUpdated:i,totalUpdates:o,totalRejectedUpdates:c,publicAddress:e.dataplane.networking.ingress.publicAddress||null,publicPort:e.dataplane.networking.ingress.publicPort||null,dpVersion:d,envoyVersion:p,type:h}),t.sortEntities(a),a})).catch((function(t){console.error(t)}))},s=function(){return i().then((function(a){var i=function(){var e=a;return"total"in e?0!==e.total&&e.items&&e.items.length>0?t.sortEntities(e.items):null:e};if(i()){a.next?(t.next=Object(l["e"])(a.next),t.hasNext=!0):t.hasNext=!1;var s=[],o=n?i():i()[0];t.firstEntity=o.name,t.getEntity(o),t.$store.dispatch("updateSelectedTableRow",t.firstEntity),n&&n.length&&e&&e.length?r(e,n,s):i().forEach((function(t){r(t.mesh,t.name,s)})),t.tableData.data=s,t.tableDataIsEmpty=!1,t.isEmpty=!1}else t.tableData.data=[],t.tableDataIsEmpty=!0,t.isEmpty=!0,t.getEntity(null)})).catch((function(e){t.hasError=!0,t.isEmpty=!0,console.error(e)})).finally((function(){setTimeout((function(){t.isLoading=!1}),"500")}))};s()},getEntity:function(t){var e=this;this.entityIsLoading=!0,this.entityIsEmpty=!1;var n=this.$route.params.mesh;if(t&&null!==t){var a="all"===n?t.mesh:n;return this.$api.getDataplaneOverviewFromMesh(a,t.name).then((function(t){if(Object(u["b"])(t)){var n=["type","name","mesh"],a=function(){var e=Object(r["a"])(regeneratorRuntime.mark((function e(){return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:return e.prev=0,e.abrupt("return",Object(u["d"])(Object(u["b"])(t),Object(u["c"])(t)));case 4:e.prev=4,e.t0=e["catch"](0),console.error(e.t0);case 7:case"end":return e.stop()}}),e,null,[[0,4]])})));return function(){return e.apply(this,arguments)}}(),i=function(){var e=Object(r["a"])(regeneratorRuntime.mark((function e(){return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:return e.t0=s["a"],e.t1=Object(s["a"])({},Object(l["f"])(Object(u["b"])(t),n)),e.t2={},e.next=5,a();case 5:return e.t3=e.sent,e.t4={status:e.t3},e.t5=(0,e.t0)(e.t1,e.t2,e.t4),e.t6=Object(u["a"])(Object(u["b"])(t)),e.abrupt("return",{basicData:e.t5,tags:e.t6});case 10:case"end":return e.stop()}}),e)})));return function(){return e.apply(this,arguments)}}();i().then((function(t){e.entity=t,e.entityNamespace=t.basicData.name,e.tabGroupTitle="Mesh: ".concat(t.basicData.name),e.entityOverviewTitle="Entity Overview for ".concat(t.basicData.name)})),e.rawEntity=Object(l["k"])(Object(u["b"])(t))}else e.entity=null,e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){setTimeout((function(){e.entityIsLoading=!1}),"500")}))}setTimeout((function(){e.entityIsEmpty=!0,e.entityIsLoading=!1}),"500")}}},g=m,w=(n("9dd2"),n("2877")),E=Object(w["a"])(g,a,i,!1,null,"77dc2b84",null);e["default"]=E.exports},"9dd2":function(t,e,n){"use strict";n("bec2")},a2bf:function(t,e,n){"use strict";var a=n("e8b5"),i=n("50c4"),r=n("0366"),s=function(t,e,n,o,l,u,c,d){var p,f=l,v=0,h=!!c&&r(c,d,3);while(v<o){if(v in n){if(p=h?h(n[v],v,e):n[v],u>0&&a(p))f=s(t,e,p,i(p.length),f,u-1)-1;else{if(f>=9007199254740991)throw TypeError("Exceed the acceptable array length");t[f]=p}f++}v++}return f};t.exports=s},bb2f:function(t,e,n){var a=n("d039");t.exports=!a((function(){return Object.isExtensible(Object.preventExtensions({}))}))},bec2:function(t,e,n){},f183:function(t,e,n){var a=n("d012"),i=n("861d"),r=n("5135"),s=n("9bf2").f,o=n("90e3"),l=n("bb2f"),u=o("meta"),c=0,d=Object.isExtensible||function(){return!0},p=function(t){s(t,u,{value:{objectID:"O"+ ++c,weakData:{}}})},f=function(t,e){if(!i(t))return"symbol"==typeof t?t:("string"==typeof t?"S":"P")+t;if(!r(t,u)){if(!d(t))return"F";if(!e)return"E";p(t)}return t[u].objectID},v=function(t,e){if(!r(t,u)){if(!d(t))return!0;if(!e)return!1;p(t)}return t[u].weakData},h=function(t){return l&&b.REQUIRED&&d(t)&&!r(t,u)&&p(t),t},b=t.exports={REQUIRED:!1,fastKey:f,getWeakData:v,onFreeze:h};a[u]=!0}}]);