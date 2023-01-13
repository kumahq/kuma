import{k as T,c1 as B,c2 as F,O as q,bU as K,c4 as M,c5 as V,c6 as H,c9 as R,c8 as D,cc as n,o as r,c as y,i as u,w as s,a as l,b,j as w,e as p,bV as v,F as z,cd as S}from"./index.0a811bc4.js";import{D as W,Q as L}from"./DataOverview.caf545ad.js";import{E as G}from"./EnvoyData.b4e22844.js";import{F as P}from"./FrameSkeleton.f5b5cbae.js";import{_ as Q}from"./LabelList.vue_vue_type_style_index_0_lang.a8d582e0.js";import{M as U}from"./MultizoneInfo.f1309f80.js";import{S as j,a as X}from"./SubscriptionHeader.05bb2ae2.js";import{T as J}from"./TabsWidget.ca485468.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.f6f964d3.js";import"./ErrorBlock.cbbbb7ee.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.cab88be1.js";import"./StatusBadge.2bc5ab95.js";import"./TagList.73d98f2e.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.92596654.js";import"./_commonjsHelpers.f037b798.js";const Y={name:"ZoneIngresses",components:{AccordionItem:B,AccordionList:F,DataOverview:W,EnvoyData:G,FrameSkeleton:P,LabelList:Q,MultizoneInfo:U,SubscriptionDetails:j,SubscriptionHeader:X,TabsWidget:J,KButton:q,KCard:K},props:{selectedZoneIngressName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},data(){return{isLoading:!0,isEmpty:!1,error:null,empty_state:{title:"No Data",message:"There are no Zone Ingresses present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Ingress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],entity:{},rawData:[],pageSize:M,next:null,subscriptionsReversed:[],pageOffset:this.offset}},computed:{...V({multicluster:"config/getMulticlusterStatus"})},watch:{$route(){this.isLoading=!0,this.isEmpty=!1,this.error=null,this.init(0)}},beforeMount(){this.init(this.offset)},methods:{init(t){this.multicluster&&this.loadData(t)},tableAction(t){const i=t;this.getEntity(i)},async loadData(t){var o;this.pageOffset=t,L.set("offset",t>0?t:null),this.isLoading=!0,this.isEmpty=!1;const i=this.$route.query.ns||null,a=this.pageSize;try{const{data:e,next:c}=await this.getZoneIngressOverviews(i,a,t);this.next=c,e.length?(this.isEmpty=!1,this.rawData=e,this.getEntity({name:(o=this.selectedZoneIngressName)!=null?o:e[0].name}),this.tableData.data=e.map(m=>{const{zoneIngressInsight:h={}}=m,f=H(h);return{...m,status:f}})):(this.tableData.data=[],this.isEmpty=!0)}catch(e){e instanceof Error?this.error=e:console.error(e),this.isEmpty=!0}finally{this.isLoading=!1}},getEntity(t){var e,c;const i=["type","name"],a=this.rawData.find(m=>m.name===t.name),o=(c=(e=a==null?void 0:a.zoneIngressInsight)==null?void 0:e.subscriptions)!=null?c:[];this.subscriptionsReversed=Array.from(o).reverse(),this.entity=R(a,i),L.set("zoneIngress",this.entity.name)},async getZoneIngressOverviews(t,i,a){if(t)return{data:[await D.getZoneIngressOverview({name:t},{size:i,offset:a})],next:null};{const{items:o,next:e}=await D.getAllZoneIngressOverviews({size:i,offset:a});return{data:o!=null?o:[],next:e}}}}},$={class:"zoneingresses"},ee={class:"entity-heading"};function te(t,i,a,o,e,c){const m=n("MultizoneInfo"),h=n("KButton"),f=n("DataOverview"),k=n("LabelList"),E=n("SubscriptionHeader"),A=n("SubscriptionDetails"),O=n("AccordionItem"),x=n("AccordionList"),Z=n("KCard"),_=n("EnvoyData"),C=n("TabsWidget"),N=n("FrameSkeleton");return r(),y("div",$,[t.multicluster===!1?(r(),u(m,{key:0})):(r(),u(N,{key:1},{default:s(()=>{var I;return[l(f,{"selected-entity-name":(I=e.entity)==null?void 0:I.name,"page-size":e.pageSize,"is-loading":e.isLoading,error:e.error,"empty-state":e.empty_state,"table-data":e.tableData,"table-data-is-empty":e.isEmpty,next:e.next,"page-offset":e.pageOffset,onTableAction:c.tableAction,onLoadData:c.loadData},{additionalControls:s(()=>[t.$route.query.ns?(r(),u(h,{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"zoneingresses"}},{default:s(()=>[b(`
            View all
          `)]),_:1})):w("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","empty-state","table-data","table-data-is-empty","next","page-offset","onTableAction","onLoadData"]),b(),e.isEmpty===!1?(r(),u(C,{key:0,"has-error":e.error!==null,"is-loading":e.isLoading,tabs:e.tabs,"initial-tab-override":"overview"},{tabHeader:s(()=>[p("h1",ee,`
            Zone Ingress: `+v(e.entity.name),1)]),overview:s(()=>[l(k,null,{default:s(()=>[p("div",null,[p("ul",null,[(r(!0),y(z,null,S(e.entity,(d,g)=>(r(),y("li",{key:g},[p("h4",null,v(g),1),b(),p("p",null,v(d),1)]))),128))])])]),_:1})]),insights:s(()=>[l(Z,{"border-variant":"noBorder"},{body:s(()=>[l(x,{"initially-open":0},{default:s(()=>[(r(!0),y(z,null,S(e.subscriptionsReversed,(d,g)=>(r(),u(O,{key:g},{"accordion-header":s(()=>[l(E,{details:d},null,8,["details"])]),"accordion-content":s(()=>[l(A,{details:d,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),"xds-configuration":s(()=>[l(_,{"data-path":"xds","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-stats":s(()=>[l(_,{"data-path":"stats","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-clusters":s(()=>[l(_,{"data-path":"clusters","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),_:1},8,["has-error","is-loading","tabs"])):w("",!0)]}),_:1}))])}const fe=T(Y,[["render",te]]);export{fe as default};
