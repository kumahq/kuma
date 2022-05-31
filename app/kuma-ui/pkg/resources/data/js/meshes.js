(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["meshes"],{"0d3d":function(t,e,a){},"1d10":function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"component-frame"},[t._t("default")],2)},i=[],r={name:"FrameSkeleton"},s=r,o=(a("a948"),a("2877")),l=Object(o["a"])(s,n,i,!1,null,"666bca0e",null);e["a"]=l.exports},2201:function(t,e,a){"use strict";a("0d3d")},"23d6":function(t,e,a){"use strict";a("b91b")},"362e":function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"all-meshes"},[a("FrameSkeleton",[a("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,next:t.next},on:{tableAction:t.tableAction,loadData:function(e){return t.loadData(e)}},scopedSlots:t._u([{key:"additionalControls",fn:function(){return[a("KButton",{staticClass:"add-mesh-button",attrs:{appearance:"primary",size:"small",to:{path:"/wizard/mesh"}},nativeOn:{click:function(e){return t.onCreateClick(e)}}},[a("span",{staticClass:"custom-control-icon"},[t._v(" + ")]),t._v(" Create Mesh ")])]},proxy:!0}])}),!1===t.isEmpty?a("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"},scopedSlots:t._u([{key:"tabHeader",fn:function(){return[t.entity.basicData?a("div",[a("h3",[t._v(" Mesh: "+t._s(t.entity.basicData.name))])]):t._e()]},proxy:!0},{key:"overview",fn:function(){return[a("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[a("div",[a("ul",t._l(t.entity.basicData,(function(e,n){return a("li",{key:n},[a("h4","creationTime"===n?[t._v(" Created ")]:"modificationTime"===n?[t._v(" Last Modified ")]:[t._v(" "+t._s(n)+" ")]),a("p","creationTime"===n||"modificationTime"===n?[t._v(" "+t._s(t._f("readableDate")(e))+" "),a("em",[t._v("("+t._s(t._f("rawDate")(e))+")")])]:[t._v(" "+t._s(e)+" ")])])})),0)]),t.entity.extendedData&&t.entity.extendedData.length?a("div",[a("ul",[t._l(t.entity.extendedData,(function(e,n){return a("li",{key:n},[a("h4",[t._v(t._s(e.label))]),e.value?a("p",{staticClass:"label-cols"},[a("span",[t._v(" "+t._s(e.value.type)+" ")]),a("span",[t._v(" "+t._s(e.value.name)+" ")])]):a("KBadge",{attrs:{size:"small",appearance:"danger"}},[t._v(" Disabled ")])],1)})),a("li",[a("h4",[t._v("Locality Aware Loadbalancing")]),t.entity.localityEnabled?a("p",[a("KBadge",{attrs:{size:"small",appearance:"success"}},[t._v(" Enabled ")])],1):a("KBadge",{attrs:{size:"small",appearance:"danger"}},[t._v(" Disabled ")])],1)],2)]):t._e()])]},proxy:!0},{key:"yaml",fn:function(){return[a("YamlView",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})]},proxy:!0},{key:"resources",fn:function(){return[a("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},t._l(t.countCols,(function(e){return a("div",{key:e},[a("ul",t._l(t.counts.slice((e-1)*t.itemsPerCol,e*t.itemsPerCol),(function(e,n){return a("li",{key:n},[a("h4",[t._v(t._s(e.title))]),a("p",[t._v(t._s(t._f("formatValue")(e.value)))])])})),0)])})),0)]},proxy:!0}],null,!1,159924467)}):t._e()],1)],1)},i=[],r=(a("7db0"),a("4160"),a("b0c0"),a("4fad"),a("d3b7"),a("25f0"),a("159b"),a("d0ff")),s=(a("96cf"),a("c964")),o=a("f3f3"),l=a("2f62"),c=a("0f82"),u=a("1d3a"),f=a("6e9b"),d=a("027b"),p=a("75bb"),y=a("bc1e"),h=a("1d10"),m=a("2778"),v=a("251b"),b=a("ff9d"),g=a("0ada"),E=a("c6ec"),_={name:"Meshes",metaInfo:{title:"Meshes"},components:{FrameSkeleton:h["a"],DataOverview:m["a"],Tabs:v["a"],YamlView:b["a"],LabelList:g["a"]},filters:{formatValue:function(t){return t?t.toLocaleString("en").toString():0},readableDate:function(t){return Object(y["f"])(t)},rawDate:function(t){return Object(y["i"])(t)}},data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Meshes present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Type",key:"type"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#resources",title:"Resources"},{hash:"#yaml",title:"YAML"}],entity:{},rawEntity:{},pageSize:E["g"],next:null,itemsPerCol:3,meshInsight:Object(f["a"])()}},computed:Object(o["a"])(Object(o["a"])({},Object(l["c"])({featureFlags:"config/featureFlags"})),{},{counts:function(){var t=this.meshInsight,e=t.policies,a=t.dataplanes.total,n=Object(o["a"])(Object(o["a"])({},Object(f["b"])()),e);return[{title:"Data plane proxies",value:a},{title:E["i"].CircuitBreaker.title,value:n.CircuitBreaker.total},{title:E["i"].FaultInjection.title,value:n.FaultInjection.total},{title:E["i"].HealthCheck.title,value:n.HealthCheck.total},{title:E["i"].ProxyTemplate.title,value:n.ProxyTemplate.total},{title:E["i"].TrafficLog.title,value:n.TrafficLog.total},{title:E["i"].TrafficPermission.title,value:n.TrafficPermission.total},{title:E["i"].TrafficRoute.title,value:n.TrafficRoute.total},{title:E["i"].TrafficTrace.title,value:n.TrafficTrace.total},{title:E["i"].RateLimit.title,value:n.RateLimit.total},{title:E["i"].Retry.title,value:n.Retry.total},{title:E["i"].Timeout.title,value:n.Timeout.total},{title:E["i"].MeshGateway.title,value:n.MeshGateway.total},{title:E["i"].MeshGatewayRoute.title,value:n.MeshGatewayRoute.total}]},countCols:function(){return Math.ceil(this.counts.length/this.itemsPerCol)}}),watch:{$route:function(t,e){this.init()}},beforeMount:function(){this.init()},methods:{init:function(){this.loadData()},onCreateClick:function(){d["a"].logger.info(p["a"].CREATE_MESH_CLICKED)},tableAction:function(t){var e=t;this.getEntity(e)},loadData:function(){var t=arguments,e=this;return Object(s["a"])(regeneratorRuntime.mark((function a(){var n,i,s,o,l,f;return regeneratorRuntime.wrap((function(a){while(1)switch(a.prev=a.next){case 0:return n=t.length>0&&void 0!==t[0]?t[0]:"0",e.isLoading=!0,e.isEmpty=!1,i=e.$route.params.mesh,"all"!==i&&(s=e.$route.params.mesh),a.prev=5,a.next=8,Object(u["a"])({getSingleEntity:c["a"].getMesh.bind(c["a"]),getAllEntities:c["a"].getAllMeshes.bind(c["a"]),size:e.pageSize,offset:n,query:s});case 8:o=a.sent,l=o.data,f=o.next,e.next=f,l.length?(e.tableData.data=Object(r["a"])(l),e.tableDataIsEmpty=!1,e.isEmpty=!1,e.getEntity({name:l[0].name})):(e.tableData.data=[],e.tableDataIsEmpty=!0,e.isEmpty=!0,e.entityIsEmpty=!0),a.next=20;break;case 15:a.prev=15,a.t0=a["catch"](5),e.hasError=!0,e.isEmpty=!0,console.error(a.t0);case 20:return a.prev=20,e.isLoading=!1,a.finish(20);case 23:case"end":return a.stop()}}),a,null,[[5,15,20,23]])})))()},getEntity:function(t){var e=this;if(this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1,t&&null!==t)return c["a"].getMesh({name:t.name}).then((function(a){if(a){c["a"].getMeshInsights({name:t.name}).then((function(t){e.meshInsight=t}));var n=Object(y["d"])(a,["type","name"]),i=function(){var t=Object.entries(Object(y["d"])(a,["mtls","logging","metrics","tracing"])),e=[];return t.forEach((function(t){var a=t[0],n=t[1]||null;if(n&&n.enabledBackend){var i=n.enabledBackend,r=n.backends.find((function(t){return t.name===i}));r&&e.push({label:a,value:{type:r.type,name:r.name}})}else if(n&&n.defaultBackend){var s=n.defaultBackend,o=n.backends.find((function(t){return t.name===s}));o&&e.push({label:a,value:{type:o.type,name:o.name}})}else if(n&&n.backends){var l=n.backends[0];l&&e.push({label:a,value:{type:l.type,name:l.name}})}else e.push({label:a,value:null})})),e},r=function(){var t=a.routing;return t&&t.localityAwareLoadBalancing};e.entity={basicData:n,extendedData:i(),localityEnabled:r()},e.rawEntity=Object(y["j"])(a)}else e.entity={},e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){setTimeout((function(){e.entityIsLoading=!1}),"500")}));setTimeout((function(){e.entityIsEmpty=!0,e.entityIsLoading=!1}),"500")}}},x=_,k=(a("2201"),a("2877")),C=Object(k["a"])(x,n,i,!1,null,"211462a3",null);e["default"]=C.exports},"536d":function(t,e,a){"use strict";a("66ad")},"62e5":function(t,e){var a;a=function(){function t(t,e){var a,n,i,r,s,o,l,c,u;null==e&&(e=""),i="",s=t.length,o=null,n=0,r=0;while(r<s){if(a=t.charAt(r),"\\"===a)i+=t.slice(r,+(r+1)+1||9e9),r++;else if("("===a)if(r<s-2)if(c=t.slice(r,+(r+2)+1||9e9),"(?:"===c)r+=2,i+=c;else if("(?<"===c){n++,r+=2,l="";while(r+1<s){if(u=t.charAt(r+1),">"===u){i+="(",r++,l.length>0&&(null==o&&(o={}),o[l]=n);break}l+=u,r++}}else i+=a,n++;else i+=a;else i+=a;r++}this.rawRegex=t,this.cleanedRegex=i,this.regex=new RegExp(this.cleanedRegex,"g"+e.replace("g","")),this.mapping=o}return t.prototype.regex=null,t.prototype.rawRegex=null,t.prototype.cleanedRegex=null,t.prototype.mapping=null,t.prototype.exec=function(t){var e,a,n,i;if(this.regex.lastIndex=0,a=this.regex.exec(t),null==a)return null;if(null!=this.mapping)for(n in i=this.mapping,i)e=i[n],a[n]=a[e];return a},t.prototype.test=function(t){return this.regex.lastIndex=0,this.regex.test(t)},t.prototype.replace=function(t,e){return this.regex.lastIndex=0,t.replace(this.regex,e)},t.prototype.replaceAll=function(t,e,a){var n;null==a&&(a=0),this.regex.lastIndex=0,n=0;while(this.regex.test(t)&&(0===a||n<a))this.regex.lastIndex=0,t=t.replace(this.regex,e),n++;return[t,n]},t}(),t.exports=a},"66ad":function(t,e,a){},"6d8a":function(t,e,a){var n,i;i=a("62e5"),n=function(){var t;function e(){}return e.LIST_ESCAPEES=["\\","\\\\",'\\"','"',"\0","","","","","","","","\b","\t","\n","\v","\f","\r","","","","","","","","","","","","","","","","","","",(t=String.fromCharCode)(133),t(160),t(8232),t(8233)],e.LIST_ESCAPED=["\\\\",'\\"','\\"','\\"',"\\0","\\x01","\\x02","\\x03","\\x04","\\x05","\\x06","\\a","\\b","\\t","\\n","\\v","\\f","\\r","\\x0e","\\x0f","\\x10","\\x11","\\x12","\\x13","\\x14","\\x15","\\x16","\\x17","\\x18","\\x19","\\x1a","\\e","\\x1c","\\x1d","\\x1e","\\x1f","\\N","\\_","\\L","\\P"],e.MAPPING_ESCAPEES_TO_ESCAPED=function(){var t,a,n,i;for(n={},t=a=0,i=e.LIST_ESCAPEES.length;0<=i?a<i:a>i;t=0<=i?++a:--a)n[e.LIST_ESCAPEES[t]]=e.LIST_ESCAPED[t];return n}(),e.PATTERN_CHARACTERS_TO_ESCAPE=new i("[\\x00-\\x1f]|Â|Â |â¨|â©"),e.PATTERN_MAPPING_ESCAPEES=new i(e.LIST_ESCAPEES.join("|").split("\\").join("\\\\")),e.PATTERN_SINGLE_QUOTING=new i("[\\s'\":{}[\\],&*#?]|^[-?|<>=!%@`]"),e.requiresDoubleQuoting=function(t){return this.PATTERN_CHARACTERS_TO_ESCAPE.test(t)},e.escapeWithDoubleQuotes=function(t){var e;return e=this.PATTERN_MAPPING_ESCAPEES.replace(t,function(t){return function(e){return t.MAPPING_ESCAPEES_TO_ESCAPED[e]}}(this)),'"'+e+'"'},e.requiresSingleQuoting=function(t){return this.PATTERN_SINGLE_QUOTING.test(t)},e.escapeWithSingleQuotes=function(t){return"'"+t.replace(/'/g,"''")+"'"},e}(),t.exports=n},a948:function(t,e,a){"use strict";a("f9f3")},b91b:function(t,e,a){},e80b:function(t,e,a){var n=a("6d8a"),i="  ";function r(t){var e=typeof t;return t instanceof Array?"array":"string"==e?"string":"boolean"==e?"boolean":"number"==e?"number":"undefined"==e||null===t?"null":"hash"}function s(t,e){var a=r(t);switch(a){case"array":o(t,e);break;case"hash":l(t,e);break;case"string":u(t,e);break;case"null":e.push("null");break;case"number":e.push(t.toString());break;case"boolean":e.push(t?"true":"false");break}}function o(t,e){for(var a=0;a<t.length;a++){var n=t[a],r=[];s(n,r);for(var o=0;o<r.length;o++)e.push((0==o?"- ":i)+r[o])}}function l(t,e){for(var a in t){var n=[];if(t.hasOwnProperty(a)){var o=t[a];s(o,n);var l=r(o);if("string"==l||"null"==l||"number"==l||"boolean"==l)e.push(c(a)+": "+n[0]);else{e.push(c(a)+": ");for(var u=0;u<n.length;u++)e.push(i+n[u])}}}}function c(t){return t.match(/^[\w]+$/)?t:n.requiresDoubleQuoting(t)?n.escapeWithDoubleQuotes(t):n.requiresSingleQuoting(t)?n.escapeWithSingleQuotes(t):t}function u(t,e){e.push(c(t))}var f=function(t){"string"==typeof t&&(t=JSON.parse(t));var e=[];return s(t,e),e.join("\n")};t.exports=f},f9f3:function(t,e,a){},ff9d:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"yaml-view"},[t.isReady?a("div",{staticClass:"yaml-view-content"},[t.isLoading||t.isEmpty?t._e():a("KCard",{attrs:{title:t.yamlTitle,"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[a("KTabs",{key:t.environment,attrs:{tabs:t.tabs},scopedSlots:t._u([{key:"universal",fn:function(){return[a("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function(e){var n=e.copyToClipboard;return[a("KPop",{attrs:{placement:"bottom"},scopedSlots:t._u([{key:"content",fn:function(){return[a("div",[a("p",[t._v("Entity copied to clipboard!")])])]},proxy:!0}],null,!0)},[a("KButton",{staticClass:"copy-button",attrs:{appearance:"primary",size:"small"},on:{click:function(){n(t.yamlContent.universal)}}},[t._v(" Copy Universal YAML ")])],1)]}}],null,!1,1536634960)}),a("Prism",{staticClass:"code-block",attrs:{language:"yaml",code:t.yamlContent.universal}})]},proxy:!0},{key:"kubernetes",fn:function(){return[a("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function(e){var n=e.copyToClipboard;return[a("KPop",{attrs:{placement:"bottom"},scopedSlots:t._u([{key:"content",fn:function(){return[a("div",[a("p",[t._v("Entity copied to clipboard!")])])]},proxy:!0}],null,!0)},[a("KButton",{staticClass:"copy-button",attrs:{appearance:"primary",size:"small"},on:{click:function(){n(t.yamlContent.kubernetes)}}},[t._v(" Copy Kubernetes YAML ")])],1)]}}],null,!1,2265429040)}),a("Prism",{staticClass:"code-block",attrs:{language:"yaml",code:t.yamlContent.kubernetes}})]},proxy:!0}],null,!1,1506056494),model:{value:t.activeTab.hash,callback:function(e){t.$set(t.activeTab,"hash",e)},expression:"activeTab.hash"}})]},proxy:!0}],null,!1,137880475)})],1):t._e(),!0===t.loaders?a("div",[t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),t._v(" Data Loading... ")]},proxy:!0}],null,!1,3263214496)}):t._e(),t.isEmpty&&!t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),t._v(" There is no data to display. ")]},proxy:!0}],null,!1,1612658095)}):t._e(),t.hasError?a("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")]},proxy:!0}],null,!1,822917942)}):t._e()],1):t._e()])},i=[],r=(a("caad"),a("a15b"),a("b0c0"),a("4fad"),a("ac1f"),a("2532"),a("1276"),a("f3f3")),s=a("2f62"),o=a("2ccf"),l=a.n(o),c=a("e80b"),u=a.n(c),f={name:"YamlView",components:{Prism:l.a},props:{title:{type:String,default:null},content:{type:Object,default:null},loaders:{type:Boolean,default:!0},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}},data:function(){return{tabs:[{hash:"#universal",title:"Universal"},{hash:"#kubernetes",title:"Kubernetes"}]}},computed:Object(r["a"])(Object(r["a"])({},Object(s["c"])({environment:"config/getEnvironment"})),{},{isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading},activeTab:{get:function(){var t=this.environment;return{hash:"#".concat(t),nohash:t}},set:function(t){return{hash:"#".concat(t),nohash:t}}},yamlTitle:function(){var t;return this.title?this.title:null!==(t=this.content)&&void 0!==t&&t.name?"Entity Overview for ".concat(this.content.name):"Entity Overview"},yamlContent:function(){var t=this,e=this.content,a=function(){var e={},a=Object.assign({},t.content),n=a.name,i=a.mesh,r=a.type,s=function(){var e=Object.assign({},t.content);return delete e.type,delete e.mesh,delete e.name,!!(e&&Object.entries(e).length>0)&&e};if(e.apiVersion="kuma.io/v1alpha1",e.kind=r,void 0!==i&&(e.mesh=a.mesh),null!==n&&void 0!==n&&n.includes(".")){var o=n.split("."),l=o.pop(),c=o.join(".");e.metadata={name:c,namespace:l}}else e.metadata={name:n};return s()&&(e.spec=s()),e},n={universal:u()(e),kubernetes:u()(a())};return n}})},d=f,p=(a("23d6"),a("536d"),a("2877")),y=Object(p["a"])(d,n,i,!1,null,"78c7b522",null);e["a"]=y.exports}}]);