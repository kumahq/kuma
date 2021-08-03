(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["dataplanes-gateway"],{1373:function(t,e,a){"use strict";a("99af");var n=a("e80b"),i=a.n(n);e["a"]={methods:{formatForCLI:function(t){var e=arguments.length>1&&void 0!==arguments[1]?arguments[1]:'" | kumactl apply -f -',a='echo "',n=i()(t);return"".concat(a).concat(n).concat(e)}}}},"17d4":function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"internal-services"},[a("FrameSkeleton",[a("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"display-data-table":!0,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,"table-data-function-text":"View","table-data-row":"name"},on:{tableAction:t.tableAction,reloadData:t.loadData}},[a("template",{slot:"additionalControls"},[this.$route.query.ns?a("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:"internal-services"}}},[a("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()],1),a("template",{slot:"pagination"},[a("Pagination",{attrs:{"has-previous":t.previous.length>0,"has-next":t.hasNext},on:{next:t.goToNextPage,previous:t.goToPreviousPage}})],1)],2),!1===t.isEmpty?a("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"}},[a("template",{slot:"tabHeader"},[a("div",[a("h3",[t._v(t._s(t.tabGroupTitle))])]),a("div",[a("EntityURLControl",{attrs:{url:t.shareUrl}})],1)]),a("template",{slot:"overview"},[a("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[a("div",[a("ul",t._l(t.entity,(function(e,n){return a("li",{key:n},[a("h4",[t._v(t._s(n))]),a("p",[t._v(" "+t._s(e)+" ")])])})),0)])])],1),a("template",{slot:"yaml"},[a("YamlView",{attrs:{lang:"yaml",title:t.entityOverviewTitle,"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})],1)],2):t._e()],1)],1)},i=[],s=(a("99af"),a("d81d"),a("b0c0"),a("d3b7"),a("bc1e")),r=a("6663"),o=a("b912"),l=a("1373"),u=a("1d10"),c=a("1799"),h=a("2778"),y=a("251b"),p=a("ff9d"),m=a("0ada"),d={name:"InternalServices",metaInfo:{title:"Internal Services"},components:{EntityURLControl:r["a"],FrameSkeleton:u["a"],Pagination:c["a"],DataOverview:h["a"],Tabs:y["a"],YamlView:p["a"],LabelList:m["a"]},mixins:[l["a"],o["a"]],data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Internal Services present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Data plane proxies: Online / Total",key:"totalOnline"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#yaml",title:"YAML"}],entity:[],rawEntity:null,firstEntity:null,pageSize:this.$pageSize,pageOffset:null,next:null,hasNext:!1,previous:[]}},computed:{tabGroupTitle:function(){var t=this.entity;return t?"Internal Services: ".concat(t.name):null},entityOverviewTitle:function(){var t=this.entity;return t?"Entity Overview for ".concat(t.name):null},formattedRawEntity:function(){var t=this.formatForCLI(this.rawEntity);return t},shareUrl:function(){var t=this,e="".concat(window.location.origin,"#"),a=this.entity,n=function(){return t.$route.query.ns?t.$route.fullPath:"".concat(e).concat(t.$route.fullPath,"?ns=").concat(a.name)};return n()}},watch:{$route:function(t,e){this.init()}},beforeMount:function(){this.init()},methods:{init:function(){this.loadData()},goToPreviousPage:function(){this.pageOffset=this.previous.pop(),this.next=null,this.loadData()},goToNextPage:function(){this.previous.push(this.pageOffset),this.pageOffset=this.next,this.next=null,this.loadData()},tableAction:function(t){var e=t;this.getEntity(e)},loadData:function(){var t=this;this.isLoading=!0;var e=this.$route.params.mesh||null,a=this.$route.query.ns||null,n={size:this.pageSize,offset:this.pageOffset},i=function(){return"all"===e?t.$api.getAllServiceInsights(n):a&&a.length&&"all"!==e?t.$api.getServiceInsight(e,a,n):t.$api.getAllServiceInsightsFromMesh(e)},r=function(){return i().then((function(e){var n=function(){var a=e;return"total"in a?0!==a.total&&a.items&&a.items.length>0?t.sortEntities(a.items):null:a},i=n();if(n()){var r=a?i:i[0];t.firstEntity=r.name,t.getEntity(Object(s["k"])(r)),t.tableData.data=a?[i]:i,e.next?(t.next=Object(s["e"])(e.next),t.hasNext=!0):t.hasNext=!1,t.tableData.data=t.tableData.data.map((function(t){var e=t.dataplanes,a=void 0===e?{}:e,n=a.online,i=void 0===n?0:n,s=a.total,r=void 0===s?0:s;switch(t.totalOnline="".concat(i," / ").concat(r),t.status){case"online":t.status="Online";break;case"partially_degraded":t.status="Partially degraded";break;case"offline":default:t.status="Offline"}return t})),t.tableDataIsEmpty=!1,t.isEmpty=!1}else t.tableData.data=[],t.tableDataIsEmpty=!0,t.isEmpty=!0,t.getEntity(null)})).catch((function(e){t.hasError=!0,t.isEmpty=!0,console.error(e)})).finally((function(){setTimeout((function(){t.isLoading=!1}),"500")}))};r()},getEntity:function(t){var e=this;this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1;var a=this.$route.params.mesh;if(t&&null!==t){var n="all"===a?t.mesh:a;return this.$api.getServiceInsight(n,t.name).then((function(t){if(t){var a=["type","name","mesh"];e.entity=Object(s["f"])(t,a),e.rawEntity=Object(s["k"])(t)}else e.entity=null,e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){setTimeout((function(){e.entityIsLoading=!1}),"500")}))}setTimeout((function(){e.entityIsEmpty=!0,e.entityIsLoading=!1}),"500")}}},f=d,v=a("2877"),g=Object(v["a"])(f,n,i,!1,null,null,null);e["default"]=g.exports},b9b6:function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"external-services"},[a("FrameSkeleton",[a("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"display-data-table":!0,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,"table-data-function-text":"View","table-data-row":"name"},on:{tableAction:t.tableAction,reloadData:t.loadData}},[a("template",{slot:"additionalControls"},[this.$route.query.ns?a("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:"external-services"}}},[a("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()],1),a("template",{slot:"pagination"},[a("Pagination",{attrs:{"has-previous":t.previous.length>0,"has-next":t.hasNext},on:{next:t.goToNextPage,previous:t.goToPreviousPage}})],1)],2),!1===t.isEmpty?a("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"}},[a("template",{slot:"tabHeader"},[a("div",[a("h3",[t._v(t._s(t.tabGroupTitle))])]),a("div",[a("EntityURLControl",{attrs:{url:t.shareUrl}})],1)]),a("template",{slot:"overview"},[a("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[a("div",[a("ul",t._l(t.entity,(function(e,n){return a("li",{key:n},[a("h4",[t._v(t._s(n))]),a("p",[t._v(" "+t._s(e)+" ")])])})),0)])])],1),a("template",{slot:"yaml"},[a("YamlView",{attrs:{lang:"yaml",title:t.entityOverviewTitle,"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})],1)],2):t._e()],1)],1)},i=[],s=(a("99af"),a("d81d"),a("b0c0"),a("d3b7"),a("bc1e")),r=a("6663"),o=a("b912"),l=a("1373"),u=a("1d10"),c=a("1799"),h=a("2778"),y=a("251b"),p=a("ff9d"),m=a("0ada"),d={name:"ExternalServices",metaInfo:{title:"External Services"},components:{EntityURLControl:r["a"],FrameSkeleton:u["a"],Pagination:c["a"],DataOverview:h["a"],Tabs:y["a"],YamlView:p["a"],LabelList:m["a"]},mixins:[l["a"],o["a"]],data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no External Services present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Address",key:"address"},{label:"TLS",key:"tlsEnabled"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#yaml",title:"YAML"}],entity:[],rawEntity:null,firstEntity:null,pageSize:this.$pageSize,pageOffset:null,next:null,hasNext:!1,previous:[]}},computed:{tabGroupTitle:function(){var t=this.entity;return t?"External Service: ".concat(t.name):null},entityOverviewTitle:function(){var t=this.entity;return t?"Entity Overview for ".concat(t.name):null},formattedRawEntity:function(){var t=this.formatForCLI(this.rawEntity);return t},shareUrl:function(){var t=this,e="".concat(window.location.origin,"#"),a=this.entity,n=function(){return t.$route.query.ns?t.$route.fullPath:"".concat(e).concat(t.$route.fullPath,"?ns=").concat(a.name)};return n()}},watch:{$route:function(t,e){this.init()}},beforeMount:function(){this.init()},methods:{init:function(){this.loadData()},goToPreviousPage:function(){this.pageOffset=this.previous.pop(),this.next=null,this.loadData()},goToNextPage:function(){this.previous.push(this.pageOffset),this.pageOffset=this.next,this.next=null,this.loadData()},tableAction:function(t){var e=t;this.getEntity(e)},loadData:function(){var t=this;this.isLoading=!0;var e=this.$route.params.mesh||null,a=this.$route.query.ns||null,n={size:this.pageSize,offset:this.pageOffset},i=function(){return"all"===e?t.$api.getAllExternalServices(n):a&&a.length&&"all"!==e?t.$api.getExternalService(e,a,n):t.$api.getAllExternalServicesFromMesh(e)},r=function(){return i().then((function(e){var n=function(){var a=e;return"total"in a?0!==a.total&&a.items&&a.items.length>0?t.sortEntities(a.items):null:a},i=n();if(n()){var r=a?i:i[0];t.firstEntity=r.name,t.getEntity(Object(s["k"])(r)),t.tableData.data=a?[i]:i,e.next?(t.next=Object(s["e"])(e.next),t.hasNext=!0):t.hasNext=!1,t.tableData.data=t.tableData.data.map((function(t){var e=t.networking,a=void 0===e?{}:e,n=a.tls,i=void 0===n?{}:n;return t.address=a.address,t.tlsEnabled=i.enabled?"Enabled":"Disabled",t})),t.tableDataIsEmpty=!1,t.isEmpty=!1}else t.tableData.data=[],t.tableDataIsEmpty=!0,t.isEmpty=!0,t.getEntity(null)})).catch((function(e){t.hasError=!0,t.isEmpty=!0,console.error(e)})).finally((function(){setTimeout((function(){t.isLoading=!1}),"500")}))};r()},getEntity:function(t){var e=this;this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1;var a=this.$route.params.mesh;if(t&&null!==t){var n="all"===a?t.mesh:a;return this.$api.getExternalService(n,t.name).then((function(t){if(t){var a=["type","name","mesh"];e.entity=Object(s["f"])(t,a),e.rawEntity=Object(s["k"])(t)}else e.entity=null,e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){setTimeout((function(){e.entityIsLoading=!1}),"500")}))}setTimeout((function(){e.entityIsEmpty=!0,e.entityIsLoading=!1}),"500")}}},f=d,v=a("2877"),g=Object(v["a"])(f,n,i,!1,null,null,null);e["default"]=g.exports},ba8f:function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"gateway-dataplanes"},[a("Dataplanes",t._b({},"Dataplanes",t.$data,!1))],1)},i=[],s=a("85e6"),r={name:"GatewayDataplanes",metaInfo:{title:"Gateway Data plane proxies"},components:{Dataplanes:s["a"]},data:function(){return{dataplaneApiParams:{gateway:!0},emptyStateMsg:"There are no Gateway data plane proxies present."}}},o=r,l=a("2877"),u=Object(l["a"])(o,n,i,!1,null,null,null);e["default"]=u.exports}}]);