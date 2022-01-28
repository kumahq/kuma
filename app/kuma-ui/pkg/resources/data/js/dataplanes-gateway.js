(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["dataplanes-gateway"],{1373:function(t,e,a){"use strict";a("99af");var n=a("e80b"),i=a.n(n);e["a"]={methods:{formatForCLI:function(t){var e=arguments.length>1&&void 0!==arguments[1]?arguments[1]:'" | kumactl apply -f -',a='echo "',n=i()(t);return"".concat(a).concat(n).concat(e)}}}},"17d4":function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("Services",{attrs:{"route-name":"internal-services",name:"Internal Services","tab-headers":t.tabHeaders}})},i=[],s=a("a100"),r={name:"InternalServices",metaInfo:{title:"Internal Services"},components:{Services:s["a"]},data:function(){return{tabHeaders:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Data plane proxies: Online / Total",key:"totalOnline"}]}}},l=r,o=a("2877"),c=Object(o["a"])(l,n,i,!1,null,null,null);e["default"]=c.exports},a100:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",[a("FrameSkeleton",[a("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,next:t.next},on:{tableAction:t.tableAction,loadData:function(e){return t.loadData(e)}},scopedSlots:t._u([{key:"additionalControls",fn:function(){return[t.$route.query.ns?a("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:t.routeName}}},[a("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()]},proxy:!0}])},[t._v(" > ")]),!1===t.isEmpty?a("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"},scopedSlots:t._u([{key:"tabHeader",fn:function(){return[a("div",[a("h3",[t._v(t._s(t.name)+": "+t._s(t.entity.name))])]),a("div",[a("EntityURLControl",{attrs:{name:t.entity.name,mesh:t.entity.mesh}})],1)]},proxy:!0},{key:"overview",fn:function(){return[a("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[a("div",[a("ul",t._l(t.entity,(function(e,n){return a("li",{key:n},[a("h4",[t._v(t._s(n))]),a("p",[t._v(" "+t._s(e)+" ")])])})),0)])])]},proxy:!0},{key:"yaml",fn:function(){return[a("YamlView",{attrs:{lang:"yaml","has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})]},proxy:!0}],null,!1,1914485537)}):t._e()],1)],1)},i=[],s=(a("99af"),a("d81d"),a("b0c0"),a("d3b7"),a("bc1e")),r=a("0f82"),l=a("6663"),o={methods:{sortEntities:function(t){var e=t.sort((function(t,e){return t.name>e.name||t.name===e.name&&t.mesh>e.mesh?1:-1}));return e}}},c=a("1373"),u=a("1d10"),m=a("2778"),d=a("251b"),h=a("ff9d"),y=a("0ada"),f=a("c6ec"),p={name:"Services",components:{EntityURLControl:l["a"],FrameSkeleton:u["a"],DataOverview:m["a"],Tabs:d["a"],YamlView:h["a"],LabelList:y["a"]},mixins:[c["a"],o],props:{routeName:{type:String,required:!0},name:{type:String,default:""},tabHeaders:{type:Array,required:!0}},data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are not ".concat(this.name," present.")},tableData:{headers:this.tabHeaders,data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#yaml",title:"YAML"}],entity:{},rawEntity:{},pageSize:f["h"],next:null}},computed:{formattedRawEntity:function(){var t=this.formatForCLI(this.rawEntity);return t}},watch:{$route:function(t,e){this.init()}},beforeMount:function(){this.init()},methods:{getAllServices:function(t){return"Internal Services"===this.name?r["a"].getAllServiceInsights(t):r["a"].getAllExternalServices(t)},getService:function(t,e,a){return"Internal Services"===this.name?r["a"].getServiceInsight({mesh:t,name:e},a):r["a"].getExternalService({mesh:t,name:e},a)},getServiceFromMesh:function(t){return"Internal Services"===this.name?r["a"].getAllServiceInsightsFromMesh({mesh:t}):r["a"].getAllExternalServicesFromMesh({mesh:t})},parseData:function(t){if("Internal Services"===this.name){var e=t.dataplanes,a=void 0===e?{}:e,n=a.online,i=void 0===n?0:n,s=a.total,r=void 0===s?0:s;switch(t.totalOnline="".concat(i," / ").concat(r),t.status){case"online":t.status=f["f"];break;case"partially_degraded":t.status=f["i"];break;case"offline":default:t.status=f["e"]}return t}var l=t.networking,o=void 0===l?{}:l,c=o.tls,u=void 0===c?{}:c;return t.address=o.address,t.tlsEnabled=u.enabled?"Enabled":"Disabled",t},init:function(){this.loadData()},goToPreviousPage:function(){this.pageOffset=this.previous.pop(),this.next=null,this.loadData()},goToNextPage:function(){this.previous.push(this.pageOffset),this.pageOffset=this.next,this.next=null,this.loadData()},tableAction:function(t){var e=t;this.getEntity(e)},loadData:function(){var t=this,e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:"0";this.isLoading=!0;var a=this.$route.params.mesh||null,n=this.$route.query.ns||null,i={size:this.pageSize,offset:e},r=function(){return"all"===a?t.getAllServices(i):n&&n.length&&"all"!==a?t.getService(a,n,i):t.getServiceFromMesh(a)},l=function(){return r().then((function(e){var a=function(){var a=e;return"total"in a?0!==a.total&&a.items&&a.items.length>0?t.sortEntities(a.items):null:a},i=a();if(a()){var r=n?i:i[0];t.getEntity(Object(s["j"])(r)),t.tableData.data=n?[i]:i,t.next=Boolean(e.next),t.tableData.data=t.tableData.data.map(t.parseData),t.tableDataIsEmpty=!1,t.isEmpty=!1}else t.tableData.data=[],t.tableDataIsEmpty=!0,t.isEmpty=!0,t.getEntity(null)})).catch((function(e){t.hasError=!0,t.isEmpty=!0,console.error(e)})).finally((function(){setTimeout((function(){t.isLoading=!1}),"500")}))};l()},getEntity:function(t){var e=this;this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1;var a=this.$route.params.mesh;if(t&&null!==t){var n="all"===a?t.mesh:a;return this.getService(n,t.name).then((function(t){if(t){var a=["type","name","mesh"];e.entity=Object(s["d"])(t,a),e.rawEntity=Object(s["j"])(t)}else e.entity={},e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){setTimeout((function(){e.entityIsLoading=!1}),"500")}))}setTimeout((function(){e.entityIsEmpty=!0,e.entityIsLoading=!1}),"500")}}},v=p,b=a("2877"),g=Object(b["a"])(v,n,i,!1,null,null,null);e["a"]=g.exports},b9b6:function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("Services",{attrs:{"route-name":"external-services",name:"External Services","tab-headers":t.tabHeaders}})},i=[],s=a("a100"),r={name:"ExternalServices",metaInfo:{title:"External Services"},components:{Services:s["a"]},data:function(){return{tabHeaders:[{key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Address",key:"address"},{label:"TLS",key:"tlsEnabled"}]}}},l=r,o=a("2877"),c=Object(o["a"])(l,n,i,!1,null,null,null);e["default"]=c.exports},ba8f:function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"gateway-dataplanes"},[a("Dataplanes",t._b({},"Dataplanes",t.$data,!1))],1)},i=[],s=a("85e6"),r={name:"GatewayDataplanes",metaInfo:{title:"Gateway Data plane proxies"},components:{Dataplanes:s["a"]},data:function(){return{dataplaneApiParams:{gateway:!0},emptyStateMsg:"There are no Gateway data plane proxies present."}}},l=r,o=a("2877"),c=Object(o["a"])(l,n,i,!1,null,null,null);e["default"]=c.exports}}]);