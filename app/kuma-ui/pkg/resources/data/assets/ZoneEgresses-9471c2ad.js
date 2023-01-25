import{u as $}from"./vue-router-573afc44.js";import{P as M,D as Q}from"./kongponents.es-3df60cd6.js";import{d as R,e as U,k as N}from"./store-2fff246d.js";import{b as q}from"./constants-31fdaf55.js";import{Q as B}from"./QueryParameter-70743f73.js";import{a as G,A as H}from"./AccordionItem-5cf952b5.js";import{D as W}from"./DataOverview-66f4d3ae.js";import{E as _}from"./EnvoyData-b1ca0b5a.js";import{F as X}from"./FrameSkeleton-256a9a83.js";import{_ as Y}from"./LabelList.vue_vue_type_style_index_0_lang-f4e0fb40.js";import{_ as j,S as J}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-875d835a.js";import{T as K}from"./TabsWidget-143f1d2e.js";import{d as ee,r as l,y as ae,_ as te,h as d,e as r,w as a,o as u,u as b,a as E,f as z,b as w,g as p,t as k,F as I,v as L}from"./runtime-dom.esm-bundler-91b41870.js";import"./vuex.esm-bundler-df5bd11e.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./datadogLogEvents-4578cfa7.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-4047971f.js";import"./ErrorBlock-a792d9a1.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-d3176fee.js";import"./StatusBadge-81464ebd.js";import"./TagList-91d1133a.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-525d6c39.js";import"./_commonjsHelpers-87174ba5.js";const se={class:"zoneegresses"},ne={class:"entity-heading"},re={key:0},Te=ee({__name:"ZoneEgresses",props:{selectedZoneEgressName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(O){const f=O,C={title:"No Data",message:"There are no Zone Egresses present."},F=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Egress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],g=$(),c=l(!0),m=l(!1),v=l(null),y=l({headers:[{label:"Actions",key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]}),i=l(null),D=l([]),S=l(null),x=l([]),A=l(f.offset);ae(()=>g.params.mesh,function(){g.name==="zoneegresses"&&(c.value=!0,m.value=!1,v.value=null,h(0))}),te(function(){h(f.offset)});async function h(t){A.value=t,B.set("offset",t>0?t:null),c.value=!0,m.value=!1;const s=g.query.ns||null,o=q;try{const{data:e,next:n}=await P(s,o,t);S.value=n,e.length?(m.value=!1,D.value=e,T({name:f.selectedZoneEgressName??e[0].name}),y.value.data=e.map(Z=>{const V=R(Z.zoneEgressInsight??{});return{...Z,status:V}})):(y.value.data=[],m.value=!0)}catch(e){e instanceof Error?v.value=e:console.error(e),m.value=!0}finally{c.value=!1}}function T({name:t}){var e;const s=D.value.find(n=>n.name===t),o=((e=s==null?void 0:s.zoneEgressInsight)==null?void 0:e.subscriptions)??[];x.value=Array.from(o).reverse(),i.value=U(s,["type","name"]),B.set("zoneEgress",t)}async function P(t,s,o){if(t)return{data:[await N.getZoneEgressOverview({name:t},{size:s,offset:o})],next:null};{const{items:e,next:n}=await N.getAllZoneEgressOverviews({size:s,offset:o});return{data:e??[],next:n}}}return(t,s)=>(u(),d("div",se,[r(X,null,{default:a(()=>{var o;return[r(W,{"selected-entity-name":(o=i.value)==null?void 0:o.name,"page-size":b(q),"is-loading":c.value,error:v.value,"empty-state":C,"table-data":y.value,"table-data-is-empty":m.value,next:S.value,"page-offset":A.value,onTableAction:T,onLoadData:h},{additionalControls:a(()=>[t.$route.query.ns?(u(),E(b(M),{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"zoneegresses"}},{default:a(()=>[z(`
            View all
          `)]),_:1})):w("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","next","page-offset"]),z(),m.value===!1&&i.value!==null?(u(),E(K,{key:0,"has-error":v.value!==null,"is-loading":c.value,tabs:F},{tabHeader:a(()=>[p("h1",ne,`
            Zone Egress: `+k(i.value.name),1)]),overview:a(()=>[r(Y,null,{default:a(()=>[p("div",null,[p("ul",null,[(u(!0),d(I,null,L(i.value,(e,n)=>(u(),d("li",{key:n},[e?(u(),d("h4",re,k(n),1)):w("",!0),z(),p("p",null,k(e),1)]))),128))])])]),_:1})]),insights:a(()=>[r(b(Q),{"border-variant":"noBorder"},{body:a(()=>[r(G,{"initially-open":0},{default:a(()=>[(u(!0),d(I,null,L(x.value,(e,n)=>(u(),E(H,{key:n},{"accordion-header":a(()=>[r(j,{details:e},null,8,["details"])]),"accordion-content":a(()=>[r(J,{details:e,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),"xds-configuration":a(()=>[r(_,{"data-path":"xds","zone-egress-name":i.value.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),"envoy-stats":a(()=>[r(_,{"data-path":"stats","zone-egress-name":i.value.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),"envoy-clusters":a(()=>[r(_,{"data-path":"clusters","zone-egress-name":i.value.name,"query-key":"envoy-data-zone-egress"},null,8,["zone-egress-name"])]),_:1},8,["has-error","is-loading"])):w("",!0)]}),_:1})]))}});export{Te as default};
