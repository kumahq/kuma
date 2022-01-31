(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["traffic-permissions"],{"42f1":function(t,e,n){"use strict";n.r(e);var a=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"traffic-permissions relative"},[n("DocumentationLink",{attrs:{href:t.docsURL}}),t.securityWarning?n("div",{staticClass:"mb-4"},[n("KAlert",{attrs:{appearance:"warning"},scopedSlots:t._u([{key:"alertMessage",fn:function(){return[n("div",{staticClass:"alert-content"},[n("p",[n("strong",[t._v("All traffic is allowed:")]),t._v(" All service traffic is enabled on this Mesh by default because Mutual TLS is not enabled. Traffic Permissions are currently being ignored by the "),n("strong",[t._v(t._s(t.$route.params.mesh))]),t._v(" Mesh because Mutual TLS is not enabled. You can still create and edit Traffic Permissions, but they will go into effect only when Mutual TLS is enabled on the Mesh. ")])])]},proxy:!0}],null,!1,484362687)})],1):t._e(),n("FrameSkeleton",[n("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,next:t.next},on:{tableAction:t.tableAction,loadData:function(e){return t.loadData(e)}},scopedSlots:t._u([{key:"additionalControls",fn:function(){return[t.$route.query.ns?n("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:"traffic-permissions"}}},[n("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()]},proxy:!0}])}),!1===t.isEmpty?n("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"},scopedSlots:t._u([{key:"tabHeader",fn:function(){return[n("div",[n("h3",[t._v("Traffic Permission:"+t._s(t.entity.name))])]),n("div",[n("EntityURLControl",{attrs:{name:t.entity.name,mesh:t.entity.mesh}})],1)]},proxy:!0},{key:"overview",fn:function(){return[n("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[n("div",[n("ul",t._l(t.entity,(function(e,a){return n("li",{key:a},[n("h4",[t._v(t._s(a))]),n("p",[t._v(" "+t._s(e)+" ")])])})),0)])])]},proxy:!0},{key:"affected-dpps",fn:function(){return[n("PolicyConnections",{attrs:{mesh:t.rawEntity.mesh,"policy-name":t.rawEntity.name,"policy-type":"traffic-permissions"}})]},proxy:!0},{key:"yaml",fn:function(){return[n("YamlView",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})]},proxy:!0}],null,!1,1243277773)}):t._e()],1)],1)},i=[],s=(n("b0c0"),n("d3b7"),n("96cf"),n("c964")),r=n("f3f3"),o=n("2f62"),l=n("0f82"),c=n("1d3a"),u=n("bc1e"),y=n("6663"),m=n("1d10"),f=n("14eb"),d=n("2778"),p=n("251b"),h=n("ff9d"),b=n("0ada"),g=n("6524"),v=n("c6ec"),E={name:"TrafficPermissions",metaInfo:{title:"Traffic Permissions"},components:{EntityURLControl:y["a"],FrameSkeleton:m["a"],DataOverview:d["a"],Tabs:p["a"],YamlView:h["a"],LabelList:b["a"],PolicyConnections:f["a"],DocumentationLink:g["a"]},data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Traffic Permissions present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"},{hash:"#yaml",title:"YAML"}],entity:{},rawEntity:{},pageSize:v["h"],next:null,securityWarning:!1}},computed:Object(r["a"])(Object(r["a"])({},Object(o["c"])({environment:"config/getEnvironment",version:"config/getVersion"})),{},{docsURL:function(){return"https://kuma.io/docs/".concat(this.version,"/policies/traffic-permissions/")}}),watch:{$route:function(t,e){this.init()}},beforeMount:function(){this.init()},methods:{init:function(){this.loadData(),this.mtlsWarning()},tableAction:function(t){var e=t;this.getEntity(e)},loadData:function(){var t=arguments,e=this;return Object(s["a"])(regeneratorRuntime.mark((function n(){var a,i,s,r,o,y,m,f;return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:return a=t.length>0&&void 0!==t[0]?t[0]:"0",e.isLoading=!0,i=e.$route.query.ns||null,s=e.$route.params.mesh||null,n.prev=4,n.next=7,Object(c["a"])({getSingleEntity:l["a"].getTrafficPermission.bind(l["a"]),getAllEntities:l["a"].getAllTrafficPermissions.bind(l["a"]),getAllEntitiesFromMesh:l["a"].getAllTrafficPermissionsFromMesh.bind(l["a"]),mesh:s,query:i,size:e.pageSize,offset:a});case 7:r=n.sent,o=r.data,y=r.next,e.next=y,o.length?(e.tableData.data=o,e.tableDataIsEmpty=!1,e.isEmpty=!1,m=["type","name","mesh"],f=o[0],e.entity=Object(u["d"])(f,m),e.rawEntity=Object(u["j"])(f)):(e.tableData.data=[],e.tableDataIsEmpty=!0,e.isEmpty=!0,e.entityIsEmpty=!0),n.next=19;break;case 14:n.prev=14,n.t0=n["catch"](4),e.hasError=!0,e.isEmpty=!0,console.error(n.t0);case 19:return n.prev=19,e.isLoading=!1,e.entityIsLoading=!1,n.finish(19);case 23:case"end":return n.stop()}}),n,null,[[4,14,19,23]])})))()},getEntity:function(t){var e=this;this.entityIsLoading=!0,this.entityIsEmpty=!1;var n=this.$route.params.mesh;if(t){var a="all"===n?t.mesh:n;return l["a"].getTrafficPermission({mesh:a,name:t.name}).then((function(t){if(t){var n=["type","name","mesh"];e.entity=Object(u["d"])(t,n),e.rawEntity=Object(u["j"])(t)}else e.entity={},e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){setTimeout((function(){e.entityIsLoading=!1}),"500")}))}setTimeout((function(){e.entityIsEmpty=!0,e.entityIsLoading=!1}),"500")},mtlsWarning:function(){var t=this,e=this.$route.params.mesh,n="all"!==e&&e;if(n)return l["a"].getMesh({name:n}).then((function(e){var n,a=e.mtls;(null===a||void 0===a||null===(n=a.enabledBackend)||void 0===n?void 0:n.length)>0?t.securityWarning=!1:t.securityWarning=!0}));this.securityWarning=!1}}},w=E,L=n("2877"),_=Object(L["a"])(w,a,i,!1,null,null,null);e["default"]=_.exports}}]);