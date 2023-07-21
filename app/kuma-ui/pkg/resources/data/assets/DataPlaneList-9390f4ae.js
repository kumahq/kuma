var Ne=Object.defineProperty;var Fe=(e,a,t)=>a in e?Ne(e,a,{enumerable:!0,configurable:!0,writable:!0,value:t}):e[a]=t;var X=(e,a,t)=>(Fe(e,typeof a!="symbol"?a+"":a,t),t);import{d as oe,q as V,c as O,s as Y,o as d,e as h,k as m,g as o,h as S,b as j,n as Ke,t as A,C as Ce,D as Le,x as ke,F as z,j as R,f as M,v as Me,A as je,p as ie,m as re,r as qe,a as B,w as P,K as ze,E as Be,P as Re,G as we}from"./index-f4ec0be6.js";import{u as ee,W as He,Q as Qe,J as Ge}from"./kongponents.es-fed304fd.js";import{C as Je}from"./ContentWrapper-67b66133.js";import{D as Ye}from"./DataOverview-706f2010.js";import{T as Ze,_ as We}from"./TextWithCopyButton-ae3a8132.js";import{f as ue,g as Xe,i as et,k as Se,w as Te,j as Pe,l as tt,e as at,x as De,m as nt,C as st,y as lt,z as ot,n as it}from"./RouteView.vue_vue_type_script_setup_true_lang-09fd82a0.js";import{a as Q,D as le}from"./DefinitionListItem-0b3f80a7.js";import{_ as rt}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-7c8e38a7.js";import{S as ut}from"./StatusBadge-0a8731e8.js";import{T as dt}from"./TagList-0012d9cb.js";import{Q as G}from"./QueryParameter-70743f73.js";const Ae=[{key:"status",label:"Status"},{key:"entity",label:"Name"},{key:"type",label:"Type"},{key:"service",label:"Service"},{key:"protocol",label:"Protocol"},{key:"zone",label:"Zone"},{key:"lastConnected",label:"Last Connected"},{key:"lastUpdated",label:"Last Updated"},{key:"totalUpdates",label:"Total Updates"},{key:"dpVersion",label:"Kuma DP version"},{key:"envoyVersion",label:"Envoy version"}],ct=["entity"],pt=Ae.filter(e=>!ct.includes(e.key)).map(e=>({tableHeaderKey:e.key,label:e.label,isChecked:!1})),xe=["status","entity","type","service","protocol","zone","lastUpdated","dpVersion"];function ft(e,a=xe){return Ae.filter(t=>a.includes(t.key)?e?!0:t.key!=="zone":!1)}function mt(e,a,t){return Math.max(a,Math.min(e,t))}const vt=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class gt{constructor(a,t){X(this,"commands");X(this,"keyMap");X(this,"boundTriggerShortcuts");this.commands=t,this.keyMap=Object.fromEntries(Object.entries(a).map(([D,i])=>[D.toLowerCase(),i])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(a){yt(a,this.keyMap,this.commands)}}function yt(e,a,t){const D=ht(e.code),i=[e.ctrlKey?"ctrl":"",e.shiftKey?"shift":"",e.altKey?"alt":"",D].filter(b=>b!=="").join("+"),u=a[i];if(!u)return;const p=t[u];p.isAllowedContext&&!p.isAllowedContext(e)||(p.shouldPreventDefaultAction&&e.preventDefault(),!(p.isDisabled&&p.isDisabled())&&p.trigger(e))}function ht(e){return vt.includes(e)?"":e.replace(/^Key/,"").toLowerCase()}function bt(e,a){const t=" "+e,D=t.matchAll(/ ([-\s\w]+):\s*/g),i=[];for(const u of Array.from(D)){if(u.index===void 0)continue;const p=_t(u[1]);if(a.length>0&&!a.includes(p))throw new Error(`Unknown field “${p}”. Known fields: ${a.join(", ")}`);const b=u.index+u[0].length,C=t.substring(b);let _;if(/^\s*["']/.test(C)){const r=C.match(/['"](.*?)['"]/);if(r!==null)_=r[1];else throw new Error(`Quote mismatch for field “${p}”.`)}else{const r=C.indexOf(" "),k=r===-1?C.length:r;_=C.substring(0,k)}_!==""&&i.push([p,_])}return i}function _t(e){return e.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(a,t)=>t===0?a:a.substring(1).toUpperCase())}const Ie=e=>(ie("data-v-2fcde9ea"),e=e(),re(),e),kt=Ie(()=>m("span",{class:"visually-hidden"},"Focus filter",-1)),wt=["for"],Tt=["id","placeholder"],Dt={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},Ct={class:"k-suggestion-list"},St={key:0,class:"k-filter-bar-error"},Pt={key:0},At=["title","data-filter-field"],xt={class:"visually-hidden"},It=Ie(()=>m("span",{class:"visually-hidden"},"Clear query",-1)),Et=oe({__name:"KFilterBar",props:{id:{type:String,required:!0},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(e,{emit:a}){const t=e,D=V(null),i=V(null),u=V(t.query),p=V([]),b=V(null),C=V(!1),_=V(-1),N=O(()=>Object.keys(t.fields)),r=O(()=>Object.entries(t.fields).slice(0,5).map(([n,c])=>({fieldName:n,...c}))),k=O(()=>N.value.length>0?`Filter by ${N.value.join(", ")}`:"Filter"),w=O(()=>t.placeholder??k.value);Y(()=>p.value,function(n,c){E(n,c)||(b.value=null,a("fields-change",{fields:n,query:u.value}))}),Y(()=>u.value,function(){u.value===""&&(b.value=null),C.value=!0});const v={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},T={submitQuery:{trigger:x,isAllowedContext(n){return i.value!==null&&n.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:L,isAllowedContext(n){return i.value!==null&&n.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:te,isAllowedContext(n){return i.value!==null&&n.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:I,isAllowedContext(n){return D.value!==null&&n.composedPath().includes(D.value)}}};function K(){const n=new gt(v,T);Me(function(){n.registerListener()}),je(function(){n.unRegisterListener()}),g(u.value)}K();function $(n){const c=n.target;g(c.value)}function x(){if(i.value instanceof HTMLInputElement)if(_.value===-1)g(i.value.value),C.value=!1;else{const n=r.value[_.value].fieldName;n&&W(i.value,n)}}function L(){J(1)}function te(){J(-1)}function J(n){_.value=mt(_.value+n,-1,r.value.length-1)}function Z(){i.value instanceof HTMLInputElement&&i.value.focus()}function ae(n){const y=n.currentTarget.getAttribute("data-filter-field");y&&i.value instanceof HTMLInputElement&&W(i.value,y)}function W(n,c){const y=u.value===""||u.value.endsWith(" ")?"":" ";u.value+=y+c+":",n.focus(),_.value=-1}function s(){u.value="",i.value instanceof HTMLInputElement&&(i.value.value="",i.value.focus(),g(""))}function l(n){n.relatedTarget===null&&I(),D.value instanceof HTMLElement&&n.relatedTarget instanceof Node&&!D.value.contains(n.relatedTarget)&&I()}function I(){C.value=!1}function g(n){b.value=null;try{const c=bt(n,N.value);c.sort((y,q)=>y[0].localeCompare(q[0])),p.value=c}catch(c){if(c instanceof Error)b.value=c,C.value=!0;else throw c}}function E(n,c){return JSON.stringify(n)===JSON.stringify(c)}return(n,c)=>(d(),h("div",{ref_key:"filterBar",ref:D,class:"k-filter-bar","data-testid":"k-filter-bar"},[m("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:Z},[kt,o(),S(j(ee),{"aria-hidden":"true",class:"k-filter-icon",color:"var(--grey-400)","data-testid":"k-filter-bar-filter-icon","hide-title":"",icon:"filter",size:"20"})]),o(),m("label",{for:`${t.id}-filter-bar-input`,class:"visually-hidden"},[Ke(n.$slots,"default",{},()=>[o(A(k.value),1)],!0)],8,wt),o(),Ce(m("input",{id:`${t.id}-filter-bar-input`,ref_key:"filterInput",ref:i,"onUpdate:modelValue":c[0]||(c[0]=y=>u.value=y),class:"k-filter-bar-input",type:"text",placeholder:w.value,"data-testid":"k-filter-bar-filter-input",onFocus:c[1]||(c[1]=y=>C.value=!0),onBlur:l,onChange:$},null,40,Tt),[[Le,u.value]]),o(),C.value?(d(),h("div",Dt,[m("div",Ct,[b.value!==null?(d(),h("p",St,A(b.value.message),1)):(d(),h("button",{key:1,class:ke(["k-submit-query-button",{"k-submit-query-button-is-selected":_.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:x},`
          Submit `+A(u.value),3)),o(),(d(!0),h(z,null,R(r.value,(y,q)=>(d(),h("div",{key:`${t.id}-${q}`,class:ke(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":_.value===q}])},[m("b",null,A(y.fieldName),1),y.description!==""?(d(),h("span",Pt,": "+A(y.description),1)):M("",!0),o(),m("button",{class:"k-apply-suggestion-button",title:`Add ${y.fieldName}:`,type:"button","data-filter-field":y.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:ae},[m("span",xt,"Add "+A(y.fieldName)+":",1),o(),S(j(ee),{"aria-hidden":"true",color:"currentColor","hide-title":"",icon:"chevronRight",size:"16"})],8,At)],2))),128))])])):M("",!0),o(),u.value!==""?(d(),h("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:s},[It,o(),S(j(ee),{"aria-hidden":"true",color:"currentColor",icon:"clear","hide-title":"",size:"20"})])):M("",!0)],512))}});const Ot=ue(Et,[["__scopeId","data-v-2fcde9ea"]]),Ee=e=>(ie("data-v-fc544ac8"),e=e(),re(),e),Ut={class:"entity-section-list"},Vt={class:"entity-title","data-testid":"data-plane-proxy-title"},$t={class:"mt-2 heading-with-icon"},Nt={key:0},Ft=Ee(()=>m("h4",null,"Insights",-1)),Kt={class:"block-list"},Lt={key:0,class:"mt-2"},Mt=Ee(()=>m("summary",null,`
                  Responses (acknowledged / sent)
                `,-1)),jt=oe({__name:"DataPlaneEntitySummary",props:{dataPlaneOverview:{type:Object,required:!0}},setup(e){const a=e,t=Xe(),{t:D}=et(),i=O(()=>({name:"data-plane-detail-view",params:{mesh:a.dataPlaneOverview.mesh,dataPlane:a.dataPlaneOverview.name}})),u=O(()=>Se(a.dataPlaneOverview.dataplane)),p=O(()=>{var k;const r=Array.from(((k=a.dataPlaneOverview.dataplaneInsight)==null?void 0:k.subscriptions)??[]);return r.reverse(),r.map(w=>{const v=w.connectTime!==void 0?Te(w.connectTime):"—",T=w.disconnectTime!==void 0?Te(w.disconnectTime):"—",K=Object.entries(w.status).filter(([$])=>!["total","lastUpdateTime"].includes($)).map(([$,x])=>{const L=`${x.responsesAcknowledged??0} / ${x.responsesSent??0}`;return{type:$.toUpperCase(),ratio:L,responsesSent:x.responsesSent??0,responsesAcknowledged:x.responsesAcknowledged??0,responsesRejected:x.responsesRejected??0}});return{subscription:w,formattedConnectDate:v,formattedDisconnectDate:T,statuses:K}})}),b=O(()=>{const{status:r}=Pe(a.dataPlaneOverview.dataplane,a.dataPlaneOverview.dataplaneInsight);return r}),C=O(()=>{const r=tt(a.dataPlaneOverview.dataplaneInsight);return r!==null?Object.entries(r).map(([k,w])=>({name:k,version:w})):[]}),_=O(()=>{var x;const r=((x=a.dataPlaneOverview.dataplaneInsight)==null?void 0:x.subscriptions)??[];if(r.length===0)return[];const k=r[r.length-1];if(!k.version)return[];const w=[],v=k.version.envoy,T=k.version.kumaDp;if(!(v.kumaDpCompatible!==void 0?v.kumaDpCompatible:!0)){const L=`Envoy ${v.version} is not supported by Kuma DP ${T.version}.`;w.push(L)}if(!(T.kumaCpCompatible!==void 0?T.kumaCpCompatible:!0)){const L=`Kuma DP ${T.version} is not supported by this Kuma control plane.`;w.push(L)}return w});async function N(r){const{mesh:k,name:w}=a.dataPlaneOverview;return await t.getDataplaneFromMesh({mesh:k,name:w},r)}return(r,k)=>{const w=qe("router-link");return d(),B(j(He),null,{body:P(()=>[m("div",Ut,[m("section",null,[m("h3",Vt,[m("span",null,[o(`
              DPP:

              `),S(Ze,{text:e.dataPlaneOverview.name},{default:P(()=>[S(w,{to:i.value},{default:P(()=>[o(A(e.dataPlaneOverview.name),1)]),_:1},8,["to"])]),_:1},8,["text"])]),o(),S(ut,{status:b.value},null,8,["status"])]),o(),S(le,{class:"mt-4"},{default:P(()=>[u.value!==null?(d(),B(Q,{key:0,term:"Tags"},{default:P(()=>[S(dt,{tags:u.value},null,8,["tags"])]),_:1})):M("",!0),o(),C.value.length>0?(d(),B(Q,{key:1,term:"Dependencies"},{default:P(()=>[m("ul",null,[(d(!0),h(z,null,R(C.value,(v,T)=>(d(),h("li",{key:T},A(v.name)+": "+A(v.version),1))),128))]),o(),_.value.length>0?(d(),h(z,{key:0},[m("h5",$t,[o(`
                  Warnings

                  `),S(j(ee),{class:"ml-1",icon:"warning",color:"var(--black-500)","secondary-color":"var(--yellow-300)",size:"20"})]),o(),(d(!0),h(z,null,R(_.value,(v,T)=>(d(),h("p",{key:T},A(v),1))),128))],64)):M("",!0)]),_:1})):M("",!0)]),_:1})]),o(),p.value.length>0?(d(),h("section",Nt,[Ft,o(),m("div",Kt,[(d(!0),h(z,null,R(p.value,(v,T)=>(d(),h("div",{key:T},[S(le,null,{default:P(()=>[S(Q,{term:"Connect time","data-testid":`data-plane-connect-time-${T}`},{default:P(()=>[o(A(v.formattedConnectDate),1)]),_:2},1032,["data-testid"]),o(),S(Q,{term:"Disconnect time","data-testid":`data-plane-disconnect-time-${T}`},{default:P(()=>[o(A(v.formattedDisconnectDate),1)]),_:2},1032,["data-testid"]),o(),S(Q,{term:j(D)("http.api.property.controlPlaneInstanceId")},{default:P(()=>[o(A(v.subscription.controlPlaneInstanceId),1)]),_:2},1032,["term"])]),_:2},1024),o(),v.statuses.length>0?(d(),h("details",Lt,[Mt,o(),S(le,null,{default:P(()=>[(d(!0),h(z,null,R(v.statuses,(K,$)=>(d(),B(Q,{key:`${T}-${$}`,term:K.type,"data-testid":`data-plane-subscription-status-${T}-${$}`},{default:P(()=>[o(A(K.ratio),1)]),_:2},1032,["term","data-testid"]))),128))]),_:2},1024)])):M("",!0)]))),128))])])):M("",!0),o(),S(rt,{id:"code-block-data-plane-summary","resource-fetcher":N,"resource-fetcher-watch-key":a.dataPlaneOverview.name,"code-max-height":"250px"},null,8,["resource-fetcher-watch-key"])])]),_:1})}}});const qt=ue(jt,[["__scopeId","data-v-fc544ac8"]]),zt=["protocol","service","zone"];function Bt(e){const a=new Map;for(const[t,D]of e){const i=zt.includes(t),u=i?"tag":t;a.has(u)||a.set(u,[]);const p=a.get(u);let b;u==="tag"?b=(i?`kuma.io/${t}:${D}`:D).replace(/\s+/g,""):b=D,p.push(b.trim())}return a}const Rt=e=>(ie("data-v-e5b4b05e"),e=e(),re(),e),Ht={key:0},Qt=Rt(()=>m("label",{for:"data-planes-type-filter",class:"mr-2"},`
              Type:
            `,-1)),Gt=["value"],Jt=["for"],Yt=["id","checked","onChange"],Zt=oe({__name:"DataPlaneList",props:{dataPlaneOverviews:{type:Array,required:!0},isLoading:{type:Boolean,required:!1,default:!1},error:{type:[Error,null],required:!1,default:null},nextUrl:{type:String,required:!1,default:null},pageOffset:{type:Number,required:!1,default:0},selectedDppName:{type:String,required:!1,default:null},isGatewayView:{type:Boolean,required:!1,default:!1},gatewayType:{type:String,required:!1,default:"true"},dppFilterFields:{type:Object,required:!0}},emits:["load-data"],setup(e,{emit:a}){const t=e,D=at(),i={true:"All",builtin:"Builtin",delegated:"Delegated"},u={title:"No Data",message:"There are no data plane proxies present."},p=V(xe),b=V({headers:[],data:[]}),C=V(G.get("filterQuery")??""),_=V(t.gatewayType),N=V({}),r=V(null),k=O(()=>D.getters["config/getMulticlusterStatus"]),w=O(()=>"tag"in t.dppFilterFields?'tag: "kuma.io/protocol: http"':"name"in t.dppFilterFields?"name: cluster":"field: value"),v=O(()=>{let s=ft(k.value,p.value);return t.isGatewayView?s=s.filter(l=>l.key!=="protocol"):s=s.filter(l=>l.key!=="type"),{data:b.value.data,headers:s}}),T=O(()=>pt.filter(s=>t.isGatewayView?s.tableHeaderKey!=="protocol":s.tableHeaderKey!=="type").filter(s=>k.value?!0:s.tableHeaderKey!=="zone").map(s=>{const l=p.value.includes(s.tableHeaderKey);return{...s,isChecked:l}}));Y(_,function(){x(0)}),Y(N,function(){x(0)}),Y(()=>t.dataPlaneOverviews,function(){J()});function K(){const s=we.get("dpVisibleTableHeaderKeys");Array.isArray(s)&&(p.value=s),J()}K();function $(s){x(s)}function x(s){const l={...N.value};"gateway"in l||(l.gateway=_.value),a("load-data",s,l)}function L(s){s.stopPropagation()}function te(s,l){const I=s.target,g=p.value.findIndex(E=>E===l);I.checked&&g===-1?p.value.push(l):!I.checked&&g>-1&&p.value.splice(g,1),we.set("dpVisibleTableHeaderKeys",Array.from(new Set(p.value)))}function J(){var s;try{b.value.data=ae(t.dataPlaneOverviews??[]),Z({name:t.selectedDppName??((s=t.dataPlaneOverviews[0])==null?void 0:s.name)})}catch(l){console.error(l)}}function Z({name:s}){s&&t.dataPlaneOverviews.length>0?(r.value=t.dataPlaneOverviews.find(l=>l.name===s)??t.dataPlaneOverviews[0],G.set(t.isGatewayView?"gateway":"dpp",r.value.name)):(r.value=null,G.set(t.isGatewayView?"gateway":"dpp",null))}function ae(s){return s.map(l=>{var pe,fe,me,ve,ge,ye;const I=l.mesh,g=l.name,E=((pe=l.dataplane.networking.gateway)==null?void 0:pe.type)||"STANDARD",n={name:E==="STANDARD"?"data-plane-detail-view":"gateway-detail-view",params:{mesh:I,dataPlane:g}},c=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],y=Se(l.dataplane).filter(f=>c.includes(f.label)),q=(fe=y.find(f=>f.label==="kuma.io/service"))==null?void 0:fe.value,Oe=(me=y.find(f=>f.label==="kuma.io/protocol"))==null?void 0:me.value,ne=(ve=y.find(f=>f.label==="kuma.io/zone"))==null?void 0:ve.value;let de;q!==void 0&&(de={name:"service-detail-view",params:{mesh:I,service:q}});let ce;ne!==void 0&&(ce={name:"zone-cp-detail-view",params:{zone:ne}});const{status:Ue}=Pe(l.dataplane,l.dataplaneInsight),Ve=((ge=l.dataplaneInsight)==null?void 0:ge.subscriptions)??[],$e={totalUpdates:0,totalRejectedUpdates:0,dpVersion:null,envoyVersion:null,selectedTime:NaN,selectedUpdateTime:NaN,version:null},U=Ve.reduce((f,F)=>{var he,be;if(F.connectTime){const _e=Date.parse(F.connectTime);(!f.selectedTime||_e>f.selectedTime)&&(f.selectedTime=_e)}const se=Date.parse(F.status.lastUpdateTime);return se&&(!f.selectedUpdateTime||se>f.selectedUpdateTime)&&(f.selectedUpdateTime=se),{totalUpdates:f.totalUpdates+parseInt(F.status.total.responsesSent??"0",10),totalRejectedUpdates:f.totalRejectedUpdates+parseInt(F.status.total.responsesRejected??"0",10),dpVersion:((he=F.version)==null?void 0:he.kumaDp.version)||f.dpVersion,envoyVersion:((be=F.version)==null?void 0:be.envoy.version)||f.envoyVersion,selectedTime:f.selectedTime,selectedUpdateTime:f.selectedUpdateTime,version:F.version||f.version}},$e),H={entity:l,detailViewRoute:n,type:E,zone:{title:ne??"—",route:ce},service:{title:q??"—",route:de},protocol:Oe??"—",status:Ue,totalUpdates:U.totalUpdates,totalRejectedUpdates:U.totalRejectedUpdates,dpVersion:U.dpVersion??"—",envoyVersion:U.envoyVersion??"—",warnings:[],unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:U.selectedUpdateTime?De(new Date(U.selectedUpdateTime).toUTCString()):"—",lastConnected:U.selectedTime?De(new Date(U.selectedTime).toUTCString()):"—",overview:l};if(U.version){const{kind:f}=nt(U.version);switch(f!==st&&H.warnings.push(f),f){case ot:H.unsupportedEnvoyVersion=!0;break;case lt:H.unsupportedKumaDPVersion=!0;break}}return k.value&&U.dpVersion&&y.find(F=>F.label===ze)&&typeof((ye=U.version)==null?void 0:ye.kumaDp.kumaCpCompatible)=="boolean"&&!U.version.kumaDp.kumaCpCompatible&&(H.warnings.push(it),H.kumaDpAndKumaCpMismatch=!0),H})}function W({fields:s,query:l}){const I=G.get("filterFields"),g=I!==null?JSON.parse(I):{},E=JSON.stringify(g),n=Object.fromEntries(Bt(s)),c=JSON.stringify(n);G.set("filterQuery",l||null),G.set("filterFields",c),E!==c&&(N.value=n)}return(s,l)=>(d(),B(Je,null,{content:P(()=>{var I;return[S(Ye,{"selected-entity-name":(I=r.value)==null?void 0:I.name,"page-size":j(Re),"is-loading":t.isLoading,error:e.error,"empty-state":u,"table-data":v.value,"table-data-is-empty":v.value.data.length===0,next:t.nextUrl!==null,"page-offset":t.pageOffset,onTableAction:Z,onLoadData:$},{additionalControls:P(()=>[S(Ot,{id:"data-plane-proxy-filter",class:"data-plane-proxy-filter",placeholder:w.value,query:C.value,fields:t.dppFilterFields,onFieldsChange:W},null,8,["placeholder","query","fields"]),o(),t.isGatewayView?(d(),h("div",Ht,[Qt,o(),Ce(m("select",{id:"data-planes-type-filter","onUpdate:modelValue":l[0]||(l[0]=g=>_.value=g),"data-testid":"data-planes-type-filter"},[(d(),h(z,null,R(i,(g,E)=>m("option",{key:E,value:E},A(g),9,Gt)),64))],512),[[Be,_.value]])])):M("",!0),o(),S(j(Qe),{label:"Columns",icon:"cogwheel","button-appearance":"outline"},{items:P(()=>[m("div",{onClick:L},[(d(!0),h(z,null,R(T.value,(g,E)=>(d(),B(j(Ge),{key:E,class:"table-header-selector-item",item:g},{default:P(()=>[m("label",{for:`data-plane-table-header-checkbox-${E}`,class:"k-checkbox table-header-selector-item-checkbox"},[m("input",{id:`data-plane-table-header-checkbox-${E}`,checked:g.isChecked,type:"checkbox",class:"k-input",onChange:n=>te(n,g.tableHeaderKey)},null,40,Yt),o(" "+A(g.label),1)],8,Jt)]),_:2},1032,["item"]))),128))])]),_:1})]),_:1},8,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","next","page-offset"])]}),sidebar:P(()=>[r.value!==null?(d(),B(qt,{key:0,"data-plane-overview":r.value},null,8,["data-plane-overview"])):(d(),B(We,{key:1}))]),_:1}))}});const da=ue(Zt,[["__scopeId","data-v-e5b4b05e"]]);export{da as D};
