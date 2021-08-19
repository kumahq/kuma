(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["traffic-routes"],{"1d3a":function(t,e,a){"use strict";a.d(e,"a",(function(){return u}));a("b0c0"),a("96cf");var n=a("c964"),i=a("d0ff"),r=a("bc1e");function s(t){return Object(i["a"])(t).sort((function(t,e){return t.name>e.name||t.name===e.name&&t.mesh>e.mesh?1:-1}))}var o=function(t){return 0!==t.total&&t.items&&t.items.length>0?s(t.items):[]};function l(t){var e=t.getSingleEntity,a=t.getAllEntities,n=t.getAllEntitiesFromMesh,i=t.mesh,r=t.query,s=t.size,o=t.offset,l={size:s,offset:o};return"all"===i?a(l):r&&r.length&&"all"!==i?e(i,r,l):n(i)}function u(t){return c.apply(this,arguments)}function c(){return c=Object(n["a"])(regeneratorRuntime.mark((function t(e){var a,n,i,s,u,c,f,p;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return a=e.getSingleEntity,n=e.getAllEntities,i=e.getAllEntitiesFromMesh,s=e.mesh,u=e.query,c=e.size,f=e.offset,t.next=3,l({getSingleEntity:a,getAllEntities:n,getAllEntitiesFromMesh:i,mesh:s,query:u,size:c,offset:f});case 3:if(p=t.sent,p){t.next=6;break}return t.abrupt("return",{data:[],next:null});case 6:return t.abrupt("return",{data:p.items?o(p):[p],next:p.next&&Object(r["d"])(p.next)});case 7:case"end":return t.stop()}}),t)}))),c.apply(this,arguments)}},6663:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"entity-url-control"},[t.shouldDisplay?a("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function(e){var n=e.copyToClipboard;return[a("KPop",{attrs:{placement:"bottom"}},[a("KButton",{attrs:{appearance:"secondary",size:"small"},on:{click:function(){n(t.url)}}},[a("KIcon",{attrs:{slot:"icon",icon:"externalLink"},slot:"icon"}),t._v(" "+t._s(t.copyButtonText)+" ")],1),a("div",{attrs:{slot:"content"},slot:"content"},[a("p",[t._v(t._s(t.confirmationText))])])],1)]}}],null,!1,1603401634)}):t._e()],1)},i=[],r=a("a026"),s=r["default"].extend({name:"EntityURLControl",props:{url:{type:String,required:!0},copyButtonText:{type:String,default:"Copy URL"},confirmationText:{type:String,default:"URL copied to clipboard!"}},computed:{shouldDisplay:function(){var t=this.$route.params.mesh||null;return!(!t||"all"===t)}}}),o=s,l=a("2877"),u=Object(l["a"])(o,n,i,!1,null,null,null);e["a"]=u.exports},8897:function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"traffic-routes"},[a("FrameSkeleton",[a("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"display-data-table":!0,"table-data":t.tableData,"table-data-is-empty":t.tableDataIsEmpty,"table-data-function-text":"View","table-data-row":"name"},on:{tableAction:t.tableAction,reloadData:t.loadData}},[a("template",{slot:"additionalControls"},[this.$route.query.ns?a("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:"traffic-routes"}}},[a("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()],1),a("template",{slot:"pagination"},[a("Pagination",{attrs:{"has-previous":t.previous.length>0,"has-next":t.hasNext},on:{next:t.goToNextPage,previous:t.goToPreviousPage}})],1)],2),!1===t.isEmpty?a("Tabs",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"}},[a("template",{slot:"tabHeader"},[a("div",[a("h3",[t._v(t._s(t.tabGroupTitle))])]),a("div",[a("EntityURLControl",{attrs:{url:t.shareUrl}})],1)]),a("template",{slot:"overview"},[a("LabelList",{attrs:{"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[a("div",[a("ul",t._l(t.entity,(function(e,n){return a("li",{key:n},[a("h4",[t._v(t._s(n))]),a("p",[t._v(" "+t._s(e)+" ")])])})),0)])])],1),a("template",{slot:"yaml"},[a("YamlView",{attrs:{title:t.entityOverviewTitle,"has-error":t.entityHasError,"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})],1)],2):t._e()],1)],1)},i=[],r=(a("99af"),a("b0c0"),a("d3b7"),a("96cf"),a("c964")),s=a("bc1e"),o=a("0f82"),l=a("1d3a"),u=a("6663"),c=a("1d10"),f=a("1799"),p=a("2778"),y=a("251b"),m=a("ff9d"),h=a("0ada"),d=a("c6ec"),g={name:"TrafficRoutes",metaInfo:{title:"Traffic Routes"},components:{EntityURLControl:u["a"],FrameSkeleton:c["a"],Pagination:f["a"],DataOverview:p["a"],Tabs:y["a"],YamlView:m["a"],LabelList:h["a"]},data:function(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Traffic Routes present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#yaml",title:"YAML"}],entity:[],rawEntity:null,firstEntity:null,pageSize:d["b"],pageOffset:null,next:null,hasNext:!1,previous:[]}},computed:{tabGroupTitle:function(){var t=this.entity;return t?"Traffic Route: ".concat(t.name):null},entityOverviewTitle:function(){var t=this.entity;return t?"Entity Overview for ".concat(t.name):null},shareUrl:function(){var t=this,e="".concat(window.location.origin,"#"),a=this.entity,n=function(){return t.$route.query.ns?t.$route.fullPath:"".concat(e).concat(t.$route.fullPath,"?ns=").concat(a.name)};return n()}},watch:{$route:function(t,e){this.init()}},beforeMount:function(){this.init()},methods:{init:function(){this.loadData()},goToPreviousPage:function(){this.pageOffset=this.previous.pop(),this.next=null,this.loadData()},goToNextPage:function(){this.previous.push(this.pageOffset),this.pageOffset=this.next,this.next=null,this.loadData()},tableAction:function(t){var e=t;this.getEntity(e)},loadData:function(){var t=this;return Object(r["a"])(regeneratorRuntime.mark((function e(){var a,n,i,r,u,c,f;return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:return t.isLoading=!0,a=t.$route.query.ns||null,n=t.$route.params.mesh||null,e.prev=3,e.next=6,Object(l["a"])({getSingleEntity:o["a"].getTrafficRoute.bind(o["a"]),getAllEntities:o["a"].getAllTrafficRoutes.bind(o["a"]),getAllEntitiesFromMesh:o["a"].getAllTrafficRoutesFromMesh.bind(o["a"]),mesh:n,query:a,size:t.pageSize,offset:t.pageOffset});case 6:i=e.sent,r=i.data,u=i.next,t.next=u,t.hasNext=!!u,r.length?(t.tableData.data=r,t.tableDataIsEmpty=!1,t.isEmpty=!1,c=["type","name","mesh"],f=r[0],t.entity=Object(s["e"])(f,c),t.rawEntity=Object(s["j"])(f)):(t.tableData.data=[],t.tableDataIsEmpty=!0,t.isEmpty=!0,t.entityIsEmpty=!0),e.next=19;break;case 14:e.prev=14,e.t0=e["catch"](3),t.hasError=!0,t.isEmpty=!0,console.error(e.t0);case 19:return e.prev=19,t.isLoading=!1,t.entityIsLoading=!1,e.finish(19);case 23:case"end":return e.stop()}}),e,null,[[3,14,19,23]])})))()},getEntity:function(t){var e=this;this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1;var a=this.$route.params.mesh;if(t){var n="all"===a?t.mesh:a;return o["a"].getTrafficRoute(n,t.name).then((function(t){if(t){var a=["type","name","mesh"];e.entity=Object(s["e"])(t,a),e.rawEntity=Object(s["j"])(t)}else e.entity=null,e.entityIsEmpty=!0})).catch((function(t){e.entityHasError=!0,console.error(t)})).finally((function(){setTimeout((function(){e.entityIsLoading=!1}),"500")}))}setTimeout((function(){e.entityIsEmpty=!0,e.entityIsLoading=!1}),"500")}}},v=g,b=a("2877"),E=Object(b["a"])(v,n,i,!1,null,null,null);e["default"]=E.exports}}]);