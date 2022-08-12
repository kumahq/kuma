"use strict";(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[722],{60453:function(t,e,s){s.r(e),s.d(e,{default:function(){return S}});var n=function(){var t=this,e=t._self._c;return e("div",{staticClass:"zoneegresses"},[e("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.empty_state,"table-data":t.tableData,"table-data-is-empty":t.isEmpty,next:t.next},on:{tableAction:t.tableAction,loadData:function(e){return t.loadData(e)}},scopedSlots:t._u([{key:"additionalControls",fn:function(){return[t.$route.query.ns?e("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:{name:"zoneegresses"}}},[e("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()]},proxy:!0}])}),!1===t.isEmpty?e("TabsWidget",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.tabs,"initial-tab-override":"overview"},scopedSlots:t._u([{key:"tabHeader",fn:function(){return[e("div",[e("h3",[t._v(" Zone Egress: "+t._s(t.entity.name))])]),e("div",[e("EntityURLControl",{attrs:{name:t.entity.name}})],1)]},proxy:!0},{key:"overview",fn:function(){return[e("LabelList",[e("div",[e("ul",t._l(t.entity,(function(s,n){return e("li",{key:n},[s?e("h4",[t._v(" "+t._s(n)+" ")]):t._e(),e("p",[t._v(" "+t._s(s)+" ")])])})),0)])])]},proxy:!0},{key:"insights",fn:function(){return[e("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[e("AccordionList",{attrs:{"initially-open":0}},t._l(t.subscriptionsReversed,(function(s,n){return e("AccordionItem",{key:n,scopedSlots:t._u([{key:"accordion-header",fn:function(){return[e("SubscriptionHeader",{attrs:{details:s}})]},proxy:!0},{key:"accordion-content",fn:function(){return[e("SubscriptionDetails",{attrs:{details:s,"is-discovery-subscription":""}})]},proxy:!0}],null,!0)})})),1)]},proxy:!0}],null,!1,4118320068)})]},proxy:!0},{key:"xds-configuration",fn:function(){return[e("XdsConfiguration",{attrs:{"zone-egress-name":t.entity.name}})]},proxy:!0},{key:"envoy-stats",fn:function(){return[e("EnvoyStats",{attrs:{"zone-egress-name":t.entity.name}})]},proxy:!0},{key:"envoy-clusters",fn:function(){return[e("EnvoyClusters",{attrs:{"zone-egress-name":t.entity.name}})]},proxy:!0}],null,!1,2132312563)}):t._e()],1)},i=[],a=s(27361),r=s.n(a),o=s(99716),l=s(4104),u=s(70172),c=s(53419),y=s(17463),d=s(56882),h=s(87673),p=s(7001),g=s(33561),f=s(66190),m=s(65404),v=s(45689),b=s(74473),E=s(49718),_=s(64082),Z=s(46077),k={name:"ZoneEgresses",components:{EnvoyClusters:Z.Z,EnvoyStats:_.Z,DataOverview:d.Z,TabsWidget:p.Z,LabelList:g.Z,AccordionList:b.Z,AccordionItem:E.Z,SubscriptionDetails:o.Z,SubscriptionHeader:l.Z,EntityURLControl:h.Z,XdsConfiguration:f.Z},metaInfo:{title:"ZoneEgresses"},data(){return{isLoading:!0,isEmpty:!1,hasError:!1,empty_state:{title:"No Data",message:"There are no Zone Egresses present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Egress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],entity:{},rawData:[],pageSize:v.NR,next:null,subscriptionsReversed:[]}},watch:{$route(){this.init()}},beforeMount(){this.init()},methods:{init(){this.loadData()},tableAction(t){const e=t;this.getEntity(e)},async loadData(t="0"){this.isLoading=!0,this.isEmpty=!1;const e=this.$route.query.ns||null;try{const{data:s,next:n}=await(0,u.W)({getAllEntities:y.Z.getAllZoneEgressOverviews.bind(y.Z),getSingleEntity:y.Z.getZoneEgressOverview.bind(y.Z),size:this.pageSize,offset:t,query:e});this.next=n,s.length?(this.isEmpty=!1,this.rawData=s,this.getEntity({name:s[0].name}),this.tableData.data=s.map((t=>{const{zoneEgressInsight:e={}}=t;return{...t,...(0,m._I)(e)}}))):(this.tableData.data=[],this.isEmpty=!0)}catch(s){this.hasError=!0,this.isEmpty=!0,console.error(s)}finally{this.isLoading=!1}},getEntity(t){const e=["type","name"],s=this.rawData.find((e=>e.name===t.name)),n=r()(s,"zoneEgressInsight.subscriptions",[]);this.subscriptionsReversed=Array.from(n).reverse(),this.entity=(0,c.wy)(s,e)}}},x=k,w=s(1001),D=(0,w.Z)(x,n,i,!1,null,null,null),S=D.exports}}]);