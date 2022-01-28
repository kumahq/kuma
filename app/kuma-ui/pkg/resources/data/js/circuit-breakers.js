(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["circuit-breakers"],{"59e8":function(t,e,a){"use strict";a.r(e);var i=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"circuit-breakers relative"},[a("DocumentationLink",{attrs:{href:t.docsURL}}),a("FrameSkeleton",[a("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,next:t.next},on:{tableAction:t.tableAction,loadData:function(e){return t.loadData(e)}},scopedSlots:t._u([{key:"additionalControls",fn:function(){return[t.$route.query.ns?a("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:"circuit-breakers"}}},[a("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()]},proxy:!0}])},[t._v(" > ")]),!1===t.isEmpty?a("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"},scopedSlots:t._u([{key:"tabHeader",fn:function(){return[a("div",[a("h3",[t._v(" Circuit Breaker: "+t._s(t.entity.name)+" ")])]),a("div",[a("EntityURLControl",{attrs:{name:t.entity.name,mesh:t.entity.mesh}})],1)]},proxy:!0},{key:"overview",fn:function(){return[a("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[a("div",[a("ul",t._l(t.entity,(function(e,i){return a("li",{key:i},[a("h4",[t._v(t._s(i))]),a("p",[t._v(" "+t._s(e)+" ")])])})),0)])])]},proxy:!0},{key:"affected-dpps",fn:function(){return[a("PolicyConnections",{attrs:{mesh:t.rawEntity.mesh,"policy-name":t.rawEntity.name,"policy-type":"circuit-breakers"}})]},proxy:!0},{key:"yaml",fn:function(){return[a("YamlView",{attrs:{lang:"yaml","has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})]},proxy:!0}],null,!1,1064309805)}):t._e()],1)],1)},n=[],r=(a("b0c0"),a("d3b7"),a("96cf"),a("c964")),s=a("f3f3"),o=a("2f62"),c=a("bc1e"),l=a("1d3a"),y=a("0f82"),u=a("6663"),m=a("1d10"),p=a("2778"),d=a("251b"),h=a("14eb"),b=a("ff9d"),f=a("0ada"),E=a("6524"),g=a("c6ec"),v={name:"CircuitBreakers",metaInfo:{title:"Circuit Breakers"},components:{EntityURLControl:u["a"],FrameSkeleton:m["a"],DataOverview:p["a"],Tabs:d["a"],YamlView:b["a"],LabelList:f["a"],PolicyConnections:h["a"],DocumentationLink:E["a"]},data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Circuit Breakers present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"},{hash:"#yaml",title:"YAML"}],entity:{},rawEntity:{},pageSize:g["h"],next:null}},computed:Object(s["a"])(Object(s["a"])({},Object(o["c"])({version:"config/getVersion"})),{},{docsURL:function(){return"https://kuma.io/docs/".concat(this.version,"/policies/circuit-breaker/")}}),watch:{$route:function(t,e){this.init()}},beforeMount:function(){this.init()},methods:{init:function(){this.loadData()},tableAction:function(t){var e=t;this.getEntity(e)},loadData:function(){var t=arguments,e=this;return Object(r["a"])(regeneratorRuntime.mark((function a(){var i,n,r,s,o,u,m,p;return regeneratorRuntime.wrap((function(a){while(1)switch(a.prev=a.next){case 0:return i=t.length>0&&void 0!==t[0]?t[0]:"0",e.isLoading=!0,n=e.$route.query.ns||null,r=e.$route.params.mesh||null,a.prev=4,a.next=7,Object(l["a"])({getSingleEntity:y["a"].getCircuitBreaker.bind(y["a"]),getAllEntities:y["a"].getAllCircuitBreakers.bind(y["a"]),getAllEntitiesFromMesh:y["a"].getAllCircuitBreakersFromMesh.bind(y["a"]),mesh:r,query:n,size:e.pageSize,offset:i});case 7:s=a.sent,o=s.data,u=s.next,e.next=u,o.length?(e.tableData.data=o,e.tableDataIsEmpty=!1,e.isEmpty=!1,m=["type","name","mesh"],p=o[0],e.entity=Object(c["d"])(p,m),e.rawEntity=Object(c["j"])(p)):(e.tableData.data=[],e.tableDataIsEmpty=!0,e.isEmpty=!0,e.entityIsEmpty=!0),a.next=19;break;case 14:a.prev=14,a.t0=a["catch"](4),e.hasError=!0,e.isEmpty=!0,console.error(a.t0);case 19:return a.prev=19,e.isLoading=!1,e.entityIsLoading=!1,a.finish(19);case 23:case"end":return a.stop()}}),a,null,[[4,14,19,23]])})))()},getEntity:function(t){var e=this;if(this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1,t)return y["a"].getCircuitBreaker({mesh:t.mesh,name:t.name}).then((function(t){if(t){var a=["type","name","mesh"];e.entity=Object(c["d"])(t,a),e.rawEntity=Object(c["j"])(t)}else e.entity={},e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){e.entityIsLoading=!1}));this.entityIsEmpty=!0,this.entityIsLoading=!1}}},k=v,w=a("2877"),L=Object(w["a"])(k,i,n,!1,null,null,null);e["default"]=L.exports}}]);