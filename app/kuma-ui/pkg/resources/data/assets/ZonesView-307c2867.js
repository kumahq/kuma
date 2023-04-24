import{_ as te,d as ne}from"./kongponents.es-9f2d376a.js";import{d as oe,u as re,r as n,v as ie,y as le,z as F,A as G,B as ue,D as ce,I as fe,j as c,b as z,e as y,i as _,g as p,w as i,h as C,f as A,E as R,o,t as N,F as J,q as Y}from"./index-5b34b65e.js";import{_ as pe}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-2890c63d.js";import{A as ge,a as me}from"./AccordionList-afa5eac7.js";import{_ as de}from"./CodeBlock.vue_vue_type_style_index_0_lang-72cdad70.js";import{D as ve}from"./DataOverview-4a430414.js";import{_ as he}from"./LabelList.vue_vue_type_style_index_0_lang-11916613.js";import{_ as ye,S as _e}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-8f436a04.js";import{T as be}from"./TabsWidget-452618aa.js";import{_ as we}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-cbc49db8.js";import{u as ke}from"./store-76151f83.js";import{u as Ee}from"./index-c89ad052.js";import{Q as j}from"./QueryParameter-70743f73.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-a940747e.js";import"./ErrorBlock-88e8a32d.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-e1caaf97.js";import"./datadogLogEvents-302eea7b.js";import"./TagList-a04c9075.js";import"./StatusBadge-df8e2e64.js";const Se={class:"zones"},Ie={key:1,class:"kcard-stack"},ze={class:"kcard-border"},Ce={class:"kcard-border"},Ae={class:"entity-heading"},Ne={key:0},Ze={key:1},Oe={key:2},Ke=oe({__name:"ZonesView",props:{selectedZoneName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(H){const Z=H,f=Ee(),Q={title:"No Data",message:"There are no Zones present."},V=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],O=re(),T=ke(),v=n(!0),g=n(!1),b=n(null),w=n(!0),h=n(!1),k=n(!1),E=n(!1),S=n({headers:[{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Storage type",key:"storeType"},{label:"Ingress",key:"hasIngress"},{label:"Egress",key:"hasEgress"},{label:"Warnings",key:"warnings",hideLabel:!0}],data:[]}),m=n(null),L=n(null),I=n([]),x=n([]),D=n(null),B=n(Z.offset),W=n(new Set),P=n(new Set);ie(()=>O.params.mesh,function(){O.name==="zones"&&(v.value=!0,g.value=!1,w.value=!0,h.value=!1,k.value=!1,E.value=!1,b.value=null,q(0))}),le(function(){q(Z.offset)});function q(s){T.getters["config/getMulticlusterStatus"]&&$(s)}function U(){return I.value.length===0?V.filter(s=>s.hash!=="#warnings"):V}function K(s){var d;let t="-",a="",e=!0;(((d=s.zoneInsight)==null?void 0:d.subscriptions)??[]).forEach(u=>{if(u.version&&u.version.kumaCp){t=u.version.kumaCp.version;const{kumaCpGlobalCompatible:ae=!0}=u.version.kumaCp;e=ae,u.config&&(a=JSON.parse(u.config).store.type)}});const l=G(s.zoneInsight);return{...s,status:l,zoneCpVersion:t,storeType:a,hasIngress:W.value.has(s.name)?"Yes":"No",hasEgress:P.value.has(s.name)?"Yes":"No",withWarnings:!e}}function X(s){const t=new Set;s.forEach(({zoneIngress:{zone:a}})=>{t.add(a)}),W.value=t}function ee(s){const t=new Set;s.forEach(({zoneEgress:{zone:a}})=>{t.add(a)}),P.value=t}async function $(s){B.value=s,j.set("offset",s>0?s:null),v.value=!0,g.value=!1;const t=O.query.ns||null,a=R;try{const[{data:e,next:r},{items:l},{items:d}]=await Promise.all([se(t,a,s),F(f.getAllZoneIngressOverviews.bind(f)),F(f.getAllZoneEgressOverviews.bind(f))]);L.value=r,e.length?(X(l),ee(d),S.value.data=e.map(K),E.value=!1,g.value=!1,await M({name:Z.selectedZoneName??e[0].name})):(S.value.data=[],E.value=!0,g.value=!0,h.value=!0)}catch(e){e instanceof Error?b.value=e:console.error(e),g.value=!0}finally{v.value=!1}}async function M({name:s}){var t;k.value=!1,w.value=!0,h.value=!1,I.value=[];try{const a=await f.getZoneOverview({name:s}),e=((t=a.zoneInsight)==null?void 0:t.subscriptions)??[],r=G(a.zoneInsight);if(m.value={...ue(a,["type","name"]),status:r,"Authentication Type":ce(a)},j.set("zone",s),x.value=Array.from(e).reverse(),e.length>0){const l=e[e.length-1],d=l.version.kumaCp.version||"-",{kumaCpGlobalCompatible:u=!0}=l.version.kumaCp;u||I.value.push({kind:fe,payload:{zoneCpVersion:d,globalCpVersion:T.getters["config/getVersion"]}}),l.config&&(D.value=JSON.stringify(JSON.parse(l.config),null,2))}}catch(a){console.error(a),m.value=null,k.value=!0,h.value=!0}finally{w.value=!1}}async function se(s,t,a){if(s)return{data:[await f.getZoneOverview({name:s},{size:t,offset:a})],next:null};{const{items:e,next:r}=await f.getAllZoneOverviews({size:t,offset:a});return{data:e??[],next:r}}}return(s,t)=>{var a;return o(),c("div",Se,[z(T).getters["config/getMulticlusterStatus"]===!1?(o(),y(pe,{key:0})):(o(),c("div",Ie,[_("div",ze,[p(ve,{"selected-entity-name":(a=m.value)==null?void 0:a.name,"page-size":z(R),"is-loading":v.value,error:b.value,"empty-state":Q,"table-data":S.value,"table-data-is-empty":E.value,"show-warnings":S.value.data.some(e=>e.withWarnings),next:L.value,"page-offset":B.value,onTableAction:M,onLoadData:$},{additionalControls:i(()=>[s.$route.query.ns?(o(),y(z(te),{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"zones"}},{default:i(()=>[C(`
              View all
            `)]),_:1})):A("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","show-warnings","next","page-offset"])]),C(),_("div",Ce,[g.value===!1&&m.value!==null?(o(),y(be,{key:0,"has-error":b.value!==null,"is-loading":v.value,tabs:U()},{tabHeader:i(()=>[_("h1",Ae,`
              Zone: `+N(m.value.name),1)]),overview:i(()=>[p(he,{"has-error":k.value,"is-loading":w.value,"is-empty":h.value},{default:i(()=>[_("div",null,[_("ul",null,[(o(!0),c(J,null,Y(m.value,(e,r)=>(o(),c("li",{key:r},[e?(o(),c("h4",Ne,N(r),1)):A("",!0),C(),r==="status"?(o(),c("p",Ze,[p(z(ne),{appearance:e==="Offline"?"danger":"success"},{default:i(()=>[C(N(e),1)]),_:2},1032,["appearance"])])):(o(),c("p",Oe,N(e),1))]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])]),insights:i(()=>[p(me,{"initially-open":0},{default:i(()=>[(o(!0),c(J,null,Y(x.value,(e,r)=>(o(),y(ge,{key:r},{"accordion-header":i(()=>[p(ye,{details:e},null,8,["details"])]),"accordion-content":i(()=>[p(_e,{details:e},null,8,["details"])]),_:2},1024))),128))]),_:1})]),config:i(()=>[D.value?(o(),y(de,{key:0,id:"code-block-zone-config",language:"json",code:D.value,"is-searchable":"","query-key":"zone-config"},null,8,["code"])):A("",!0)]),warnings:i(()=>[p(we,{warnings:I.value},null,8,["warnings"])]),_:1},8,["has-error","is-loading","tabs"])):A("",!0)])]))])}}});export{Ke as default};
