(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["traffic-permissions"],{"249c":function(t,e,a){},"42f1":function(t,e,a){"use strict";a.r(e);var i=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"traffic-permissions"},[t.securityWarning?a("div",{staticClass:"alert-wrapper"},[a("KAlert",{attrs:{appearance:"warning"}},[a("template",{slot:"alertMessage"},[a("div",{staticClass:"alert-content"},[a("p",[a("strong",[t._v("All traffic is allowed:")]),t._v(" All service traffic is enabled on this Mesh by default because Mutual TLS is not enabled. Traffic Permissions are currently being ignored by the "),a("strong",[t._v(t._s(t.$route.params.mesh))]),t._v(" Mesh because Mutual TLS is not enabled. You can still create and edit Traffic Permissions, but they will go into effect only when Mutual TLS is enabled on the Mesh. ")])])])],2)],1):t._e(),a("FrameSkeleton",[a("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"display-data-table":!0,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,"table-data-function-text":"View","table-data-row":"name"},on:{tableAction:t.tableAction,reloadData:t.loadData}},[a("template",{slot:"additionalControls"},[this.$route.query.ns?a("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:"traffic-permissions"}}},[a("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()],1),a("template",{slot:"pagination"},[a("Pagination",{attrs:{"has-previous":t.previous.length>0,"has-next":t.hasNext},on:{next:t.goToNextPage,previous:t.goToPreviousPage}})],1)],2),!1===t.isEmpty?a("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"}},[a("template",{slot:"tabHeader"},[a("div",[a("h3",[t._v(t._s(t.tabGroupTitle))])]),a("div",[a("EntityURLControl",{attrs:{url:t.shareUrl}})],1)]),a("template",{slot:"overview"},[a("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[a("div",[a("ul",t._l(t.entity,(function(e,i){return a("li",{key:i},[a("h4",[t._v(t._s(i))]),a("p",[t._v(" "+t._s(e)+" ")])])})),0)])])],1),a("template",{slot:"yaml"},[a("YamlView",{attrs:{title:t.entityOverviewTitle,"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})],1)],2):t._e()],1)],1)},n=[],s=(a("99af"),a("b0c0"),a("d3b7"),a("f3f3")),r=a("2f62"),o=a("bc1e"),l=a("6663"),c=a("b912"),u=a("1d10"),f=a("1799"),p=a("2778"),m=a("251b"),h=a("ff9d"),y=a("0ada"),d={name:"TrafficPermissions",metaInfo:{title:"Traffic Permissions"},components:{EntityURLControl:l["a"],FrameSkeleton:u["a"],Pagination:f["a"],DataOverview:p["a"],Tabs:m["a"],YamlView:h["a"],LabelList:y["a"]},mixins:[c["a"]],data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Traffic Permissions present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#yaml",title:"YAML"}],entity:[],rawEntity:null,firstEntity:null,pageSize:this.$pageSize,pageOffset:null,next:null,hasNext:!1,previous:[],securityWarning:!1}},computed:Object(s["a"])(Object(s["a"])({},Object(r["c"])({environment:"config/getEnvironment"})),{},{tabGroupTitle:function(){var t=this.entity;return t?"Traffic Permission: ".concat(t.name):null},entityOverviewTitle:function(){var t=this.entity;return t?"Entity Overview for ".concat(t.name):null},shareUrl:function(){var t=this,e="".concat(window.location.origin,"#"),a=this.entity,i=function(){return t.$route.query.ns?t.$route.fullPath:"".concat(e).concat(t.$route.fullPath,"?ns=").concat(a.name)};return i()}}),watch:{$route:function(t,e){this.init()}},beforeMount:function(){this.init()},methods:{init:function(){this.loadData(),this.mtlsWarning()},goToPreviousPage:function(){this.pageOffset=this.previous.pop(),this.next=null,this.loadData()},goToNextPage:function(){this.previous.push(this.pageOffset),this.pageOffset=this.next,this.next=null,this.loadData()},tableAction:function(t){var e=t;this.getEntity(e)},loadData:function(){var t=this;this.isLoading=!0;var e=this.$route.params.mesh||null,a=this.$route.query.ns||null,i={size:this.pageSize,offset:this.pageOffset},n=function(){return"all"===e?t.$api.getAllTrafficPermissions(i):a&&a.length&&"all"!==e?t.$api.getTrafficPermission(e,a):t.$api.getAllTrafficPermissionsFromMesh(e)},s=function(){return n().then((function(e){var i=function(){var a=e;return"total"in a?0!==a.total&&a.items&&a.items.length>0?t.sortEntities(a.items):null:a},n=i();if(i()){var s=a?n:n[0];t.firstEntity=s.name,t.getEntity(Object(o["k"])(s)),e.next?(t.next=Object(o["e"])(e.next),t.hasNext=!0):t.hasNext=!1,t.tableData.data=a?[n]:n,t.tableDataIsEmpty=!1,t.isEmpty=!1}else t.tableData.data=[],t.tableDataIsEmpty=!0,t.isEmpty=!0,t.getEntity(null)})).catch((function(e){t.hasError=!0,t.isEmpty=!0,console.error(e)})).finally((function(){setTimeout((function(){t.isLoading=!1}),"500")}))};s()},getEntity:function(t){var e=this;this.entityIsLoading=!0,this.entityIsEmpty=!1;var a=this.$route.params.mesh;if(t&&null!==t){var i="all"===a?t.mesh:a;return this.$api.getTrafficPermission(i,t.name).then((function(t){if(t){var a=["type","name","mesh"];e.entity=Object(o["f"])(t,a),e.rawEntity=Object(o["k"])(t)}else e.entity=null,e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){setTimeout((function(){e.entityIsLoading=!1}),"500")}))}setTimeout((function(){e.entityIsEmpty=!0,e.entityIsLoading=!1}),"500")},mtlsWarning:function(){var t=this,e=this.$route.params.mesh,a="all"!==e&&e;if(a)return this.$api.getMesh(a).then((function(e){var a=function(){var a,i,n=t.environment.toLowerCase(),s=function(){return"universal"===n?!!e.mtls&&e.mtls:"kubernetes"===n&&(!!e.spec.mtls&&e.spec.mtls)};return Boolean((null===(a=s())||void 0===a||null===(i=a.enabledBackend)||void 0===i?void 0:i.length)>0)};a()?t.securityWarning=!1:t.securityWarning=!0}));this.securityWarning=!1}}},v=d,g=(a("66e0"),a("2877")),b=Object(g["a"])(v,i,n,!1,null,"7f21eb5c",null);e["default"]=b.exports},6663:function(t,e,a){"use strict";var i=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"entity-url-control"},[t.shouldDisplay?a("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function(e){var i=e.copyToClipboard;return[a("KPop",{attrs:{placement:"bottom"}},[a("KButton",{attrs:{appearance:"secondary",size:"small"},on:{click:function(){i(t.url)}}},[a("KIcon",{attrs:{slot:"icon",icon:"externalLink"},slot:"icon"}),t._v(" "+t._s(t.copyButtonText)+" ")],1),a("div",{attrs:{slot:"content"},slot:"content"},[a("p",[t._v(t._s(t.confirmationText))])])],1)]}}],null,!1,1603401634)}):t._e()],1)},n=[],s=a("a026"),r=s["default"].extend({name:"EntityURLControl",props:{url:{type:String,required:!0},copyButtonText:{type:String,default:"Copy URL"},confirmationText:{type:String,default:"URL copied to clipboard!"}},computed:{shouldDisplay:function(){var t=this.$route.params.mesh||null;return!(!t||"all"===t)}}}),o=r,l=a("2877"),c=Object(l["a"])(o,i,n,!1,null,null,null);e["a"]=c.exports},"66e0":function(t,e,a){"use strict";a("249c")}}]);