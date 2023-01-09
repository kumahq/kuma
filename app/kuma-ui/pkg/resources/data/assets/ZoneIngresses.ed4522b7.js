import{G as T,cm as B,cn as F,O as N,N as q,P as K,co as M,cp as V,x as H,k as D,h as a,o as r,i as h,c as u,w as n,a as l,b as _,A as I,j as p,t as b,F as w,n as z}from"./index.5ff69d99.js";import{D as G,p as P}from"./patchQueryParam.9388b199.js";import{E as R}from"./EnvoyData.771f4f54.js";import{F as W}from"./FrameSkeleton.48d703a0.js";import{_ as j}from"./LabelList.vue_vue_type_style_index_0_lang.6b9707d9.js";import{M as Q}from"./MultizoneInfo.e8273762.js";import{S as U,a as X}from"./SubscriptionHeader.2bdb785b.js";import{T as J}from"./TabsWidget.b5204f9d.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.815effe0.js";import"./ErrorBlock.9183b6a0.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.5eb6ebf0.js";import"./StatusBadge.c8d9a87a.js";import"./TagList.7813b18d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.21ca60cf.js";import"./_commonjsHelpers.f037b798.js";const Y={name:"ZoneIngresses",components:{AccordionItem:B,AccordionList:F,DataOverview:G,EnvoyData:R,FrameSkeleton:W,LabelList:j,MultizoneInfo:Q,SubscriptionDetails:U,SubscriptionHeader:X,TabsWidget:J,KButton:N,KCard:q},props:{offset:{type:Number,required:!1,default:0}},data(){return{isLoading:!0,isEmpty:!1,error:null,empty_state:{title:"No Data",message:"There are no Zone Ingresses present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Ingress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],entity:{},rawData:[],pageSize:K,next:null,subscriptionsReversed:[],pageOffset:this.offset}},computed:{...M({multicluster:"config/getMulticlusterStatus"})},watch:{$route(){this.isLoading=!0,this.isEmpty=!1,this.error=null,this.init(0)}},beforeMount(){this.init(this.offset)},methods:{init(t){this.multicluster&&this.loadData(t)},tableAction(t){const i=t;this.getEntity(i)},async loadData(t){this.pageOffset=t,P("offset",t>0?t:null),this.isLoading=!0,this.isEmpty=!1;const i=this.$route.query.ns||null,o=this.pageSize;try{const{data:s,next:e}=await this.getZoneIngressOverviews(i,o,t);this.next=e,s.length?(this.isEmpty=!1,this.rawData=s,this.getEntity({name:s[0].name}),this.tableData.data=s.map(c=>{const{zoneIngressInsight:m={}}=c,y=V(m);return{...c,status:y}})):(this.tableData.data=[],this.isEmpty=!0)}catch(s){s instanceof Error?this.error=s:console.error(s),this.isEmpty=!0}finally{this.isLoading=!1}},getEntity(t){var e,c;const i=["type","name"],o=this.rawData.find(m=>m.name===t.name),s=(c=(e=o==null?void 0:o.zoneIngressInsight)==null?void 0:e.subscriptions)!=null?c:[];this.subscriptionsReversed=Array.from(s).reverse(),this.entity=H(o,i)},async getZoneIngressOverviews(t,i,o){if(t)return{data:[await D.getZoneIngressOverview({name:t},{size:i,offset:o})],next:null};{const{items:s,next:e}=await D.getAllZoneIngressOverviews({size:i,offset:o});return{data:s!=null?s:[],next:e}}}}},$={class:"zoneingresses"},ee={class:"entity-heading"};function te(t,i,o,s,e,c){const m=a("MultizoneInfo"),y=a("KButton"),L=a("DataOverview"),S=a("LabelList"),k=a("SubscriptionHeader"),E=a("SubscriptionDetails"),A=a("AccordionItem"),x=a("AccordionList"),O=a("KCard"),f=a("EnvoyData"),Z=a("TabsWidget"),C=a("FrameSkeleton");return r(),h("div",$,[t.multicluster===!1?(r(),u(m,{key:0})):(r(),u(C,{key:1},{default:n(()=>{var v;return[l(L,{"selected-entity-name":(v=e.entity)==null?void 0:v.name,"page-size":e.pageSize,"is-loading":e.isLoading,error:e.error,"empty-state":e.empty_state,"table-data":e.tableData,"table-data-is-empty":e.isEmpty,next:e.next,"page-offset":e.pageOffset,onTableAction:c.tableAction,onLoadData:c.loadData},{additionalControls:n(()=>[t.$route.query.ns?(r(),u(y,{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"zoneingresses"}},{default:n(()=>[_(`
            View all
          `)]),_:1})):I("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","empty-state","table-data","table-data-is-empty","next","page-offset","onTableAction","onLoadData"]),_(),e.isEmpty===!1?(r(),u(Z,{key:0,"has-error":e.error!==null,"is-loading":e.isLoading,tabs:e.tabs,"initial-tab-override":"overview"},{tabHeader:n(()=>[p("h1",ee,`
            Zone Ingress: `+b(e.entity.name),1)]),overview:n(()=>[l(S,null,{default:n(()=>[p("div",null,[p("ul",null,[(r(!0),h(w,null,z(e.entity,(d,g)=>(r(),h("li",{key:g},[p("h4",null,b(g),1),_(),p("p",null,b(d),1)]))),128))])])]),_:1})]),insights:n(()=>[l(O,{"border-variant":"noBorder"},{body:n(()=>[l(x,{"initially-open":0},{default:n(()=>[(r(!0),h(w,null,z(e.subscriptionsReversed,(d,g)=>(r(),u(A,{key:g},{"accordion-header":n(()=>[l(k,{details:d},null,8,["details"])]),"accordion-content":n(()=>[l(E,{details:d,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),"xds-configuration":n(()=>[l(f,{"data-path":"xds","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-stats":n(()=>[l(f,{"data-path":"stats","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-clusters":n(()=>[l(f,{"data-path":"clusters","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),_:1},8,["has-error","is-loading","tabs"])):I("",!0)]}),_:1}))])}const fe=T(Y,[["render",te]]);export{fe as default};
