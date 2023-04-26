var Ke=Object.defineProperty;var Me=(e,l,t)=>l in e?Ke(e,l,{enumerable:!0,configurable:!0,writable:!0,value:t}):e[l]=t;var W=(e,l,t)=>(Me(e,typeof l!="symbol"?l+"":l,t),t);import{d as ue,r as E,c as S,v as Q,o as u,j as w,i as _,h as o,g as D,b as d,J as qe,t as P,$ as Ae,a0 as Re,Y as Te,F,q as M,f as V,k as je,a1 as ze,p as de,m as ce,L as Ee,X as De,K as xe,M as Be,x as He,e as N,w as C,u as Qe,C as Ce,G as Pe,O as Ge,Q as Ye,V as Je,W as Ze,S as We,U as Xe,a2 as et,B as tt,a3 as at}from"./index-6ef061d4.js";import{d as X,x as nt,k as st,D as lt,M as Se}from"./kongponents.es-5ca9e130.js";import{C as ot}from"./ContentWrapper-09c5378f.js";import{D as it}from"./DataOverview-273e56b2.js";import{D as K,a as re,_ as rt}from"./DefinitionListItem-97bb646e.js";import{_ as pe}from"./_plugin-vue_export-helper-c27b6911.js";import{S as ut}from"./StatusBadge-310603a4.js";import{T as dt}from"./TagList-2830199d.js";import{_ as ct}from"./YamlView.vue_vue_type_script_setup_true_lang-f94f7cbc.js";import{u as pt}from"./store-444aa12f.js";import{d as ft}from"./datadogLogEvents-302eea7b.js";import{Q as R}from"./QueryParameter-70743f73.js";const Oe=[{key:"status",label:"Status"},{key:"name",label:"DPP"},{key:"type",label:"Type"},{key:"service",label:"Service"},{key:"protocol",label:"Protocol"},{key:"zone",label:"Zone"},{key:"lastConnected",label:"Last Connected"},{key:"lastUpdated",label:"Last Updated"},{key:"totalUpdates",label:"Total Updates"},{key:"dpVersion",label:"Kuma DP version"},{key:"envoyVersion",label:"Envoy version"},{key:"details",label:"Details",hideLabel:!0}],mt=["name","details"],vt=Oe.filter(e=>!mt.includes(e.key)).map(e=>({tableHeaderKey:e.key,label:e.label,isChecked:!1})),Ie=["status","name","type","service","protocol","zone","lastUpdated","dpVersion","details"];function gt(e,l=Ie){return Oe.filter(t=>l.includes(t.key)?e?!0:t.key!=="zone":!1)}function yt(e,l,t){return Math.max(l,Math.min(e,t))}const ht=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class bt{constructor(l,t){W(this,"commands");W(this,"keyMap");W(this,"boundTriggerShortcuts");this.commands=t,this.keyMap=Object.fromEntries(Object.entries(l).map(([T,i])=>[T.toLowerCase(),i])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(l){_t(l,this.keyMap,this.commands)}}function _t(e,l,t){const T=kt(e.code),i=[e.ctrlKey?"ctrl":"",e.shiftKey?"shift":"",e.altKey?"alt":"",T].filter(g=>g!=="").join("+"),c=l[i];if(!c)return;const k=t[c];k.isAllowedContext&&!k.isAllowedContext(e)||(k.shouldPreventDefaultAction&&e.preventDefault(),!(k.isDisabled&&k.isDisabled())&&k.trigger(e))}function kt(e){return ht.includes(e)?"":e.replace(/^Key/,"").toLowerCase()}function wt(e,l){const t=" "+e,T=t.matchAll(/ ([-\s\w]+):\s*/g),i=[];for(const c of Array.from(T)){if(c.index===void 0)continue;const k=Tt(c[1]);if(l.length>0&&!l.includes(k))throw new Error(`Unknown field “${k}”. Known fields: ${l.join(", ")}`);const g=c.index+c[0].length,r=t.substring(g);let y;if(/^\s*["']/.test(r)){const p=r.match(/['"](.*?)['"]/);if(p!==null)y=p[1];else throw new Error(`Quote mismatch for field “${k}”.`)}else{const p=r.indexOf(" "),m=p===-1?r.length:p;y=r.substring(0,m)}y!==""&&i.push([k,y])}return i}function Tt(e){return e.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(l,t)=>t===0?l:l.substring(1).toUpperCase())}const Ue=e=>(de("data-v-715248af"),e=e(),ce(),e),Dt=Ue(()=>_("span",{class:"visually-hidden"},"Focus filter",-1)),Ct=["for"],Pt=["id","placeholder"],St={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},At={class:"k-suggestion-list"},Et={key:0,class:"k-filter-bar-error"},xt={key:0},Ot=["title","data-filter-field"],It={class:"visually-hidden"},Ut=Ue(()=>_("span",{class:"visually-hidden"},"Clear query",-1)),$t=ue({__name:"KFilterBar",props:{id:{type:String,required:!0},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(e,{emit:l}){const t=e,T=E(null),i=E(null),c=E(t.query),k=E([]),g=E(null),r=E(!1),y=E(-1),h=S(()=>Object.keys(t.fields)),p=S(()=>Object.entries(t.fields).slice(0,5).map(([n,s])=>({fieldName:n,...s}))),m=S(()=>h.value.length>0?`Filter by ${h.value.join(", ")}`:"Filter"),O=S(()=>t.placeholder??m.value);Q(()=>k.value,function(n,s){v(n,s)||(g.value=null,l("fields-change",{fields:n,query:c.value}))}),Q(()=>c.value,function(){c.value===""&&(g.value=null),r.value=!0});const I={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},x={submitQuery:{trigger:G,isAllowedContext(n){return i.value!==null&&n.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:te,isAllowedContext(n){return i.value!==null&&n.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:j,isAllowedContext(n){return i.value!==null&&n.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:H,isAllowedContext(n){return T.value!==null&&n.composedPath().includes(T.value)}}};function $(){const n=new bt(I,x);je(function(){n.registerListener()}),ze(function(){n.unRegisterListener()}),a(c.value)}$();function ee(n){const s=n.target;a(s.value)}function G(){if(i.value instanceof HTMLInputElement)if(y.value===-1)a(i.value.value),r.value=!1;else{const n=p.value[y.value].fieldName;n&&z(i.value,n)}}function te(){Y(1)}function j(){Y(-1)}function Y(n){y.value=yt(y.value+n,-1,p.value.length-1)}function ae(){i.value instanceof HTMLInputElement&&i.value.focus()}function ne(n){const f=n.currentTarget.getAttribute("data-filter-field");f&&i.value instanceof HTMLInputElement&&z(i.value,f)}function z(n,s){const f=c.value===""||c.value.endsWith(" ")?"":" ";c.value+=f+s+":",n.focus(),y.value=-1}function B(){c.value="",i.value instanceof HTMLInputElement&&(i.value.value="",i.value.focus(),a(""))}function se(n){n.relatedTarget===null&&H(),T.value instanceof HTMLElement&&n.relatedTarget instanceof Node&&!T.value.contains(n.relatedTarget)&&H()}function H(){r.value=!1}function a(n){g.value=null;try{const s=wt(n,h.value);s.sort((f,U)=>f[0].localeCompare(U[0])),k.value=s}catch(s){if(s instanceof Error)g.value=s,r.value=!0;else throw s}}function v(n,s){return JSON.stringify(n)===JSON.stringify(s)}return(n,s)=>(u(),w("div",{ref_key:"filterBar",ref:T,class:"k-filter-bar","data-testid":"k-filter-bar"},[_("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:ae},[Dt,o(),D(d(X),{"aria-hidden":"true",class:"k-filter-icon",color:"var(--grey-400)","data-testid":"k-filter-bar-filter-icon","hide-title":"",icon:"filter",size:"20"})]),o(),_("label",{for:`${t.id}-filter-bar-input`,class:"visually-hidden"},[qe(n.$slots,"default",{},()=>[o(P(d(m)),1)],!0)],8,Ct),o(),Ae(_("input",{id:`${t.id}-filter-bar-input`,ref_key:"filterInput",ref:i,"onUpdate:modelValue":s[0]||(s[0]=f=>c.value=f),class:"k-filter-bar-input",type:"text",placeholder:d(O),"data-testid":"k-filter-bar-filter-input",onFocus:s[1]||(s[1]=f=>r.value=!0),onBlur:se,onChange:ee},null,40,Pt),[[Re,c.value]]),o(),r.value?(u(),w("div",St,[_("div",At,[g.value!==null?(u(),w("p",Et,P(g.value.message),1)):(u(),w("button",{key:1,class:Te(["k-submit-query-button",{"k-submit-query-button-is-selected":y.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:G},`
          Submit `+P(c.value),3)),o(),(u(!0),w(F,null,M(d(p),(f,U)=>(u(),w("div",{key:`${t.id}-${U}`,class:Te(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":y.value===U}])},[_("b",null,P(f.fieldName),1),f.description!==""?(u(),w("span",xt,": "+P(f.description),1)):V("",!0),o(),_("button",{class:"k-apply-suggestion-button",title:`Add ${f.fieldName}:`,type:"button","data-filter-field":f.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:ne},[_("span",It,"Add "+P(f.fieldName)+":",1),o(),D(d(X),{"aria-hidden":"true",color:"currentColor","hide-title":"",icon:"chevronRight",size:"16"})],8,Ot)],2))),128))])])):V("",!0),o(),c.value!==""?(u(),w("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:B},[Ut,o(),D(d(X),{"aria-hidden":"true",color:"currentColor",icon:"clear","hide-title":"",size:"20"})])):V("",!0)],512))}});const Lt=pe($t,[["__scopeId","data-v-715248af"]]),$e=e=>(de("data-v-3fa16dcc"),e=e(),ce(),e),Vt={class:"entity-section-list"},Nt={class:"entity-title","data-testid":"data-plane-proxy-title"},Ft={class:"mt-2 heading-with-icon"},Kt={key:0},Mt=$e(()=>_("h4",null,"Insights",-1)),qt={class:"block-list"},Rt={key:0,class:"mt-2"},jt=$e(()=>_("summary",null,`
                  Responses (acknowledged / sent)
                `,-1)),zt={class:"config-section"},Bt=ue({__name:"DataPlaneEntitySummary",props:{dataPlaneOverview:{type:Object,required:!0}},setup(e){const l=e,t=S(()=>{const{name:r,mesh:y,dataplane:h}=l.dataPlaneOverview;return{type:"Dataplane",name:r,mesh:y,networking:h.networking}}),T=S(()=>Ee(l.dataPlaneOverview.dataplane)),i=S(()=>{var y;const r=Array.from(((y=l.dataPlaneOverview.dataplaneInsight)==null?void 0:y.subscriptions)??[]);return r.reverse(),r.map(h=>{const p=h.connectTime!==void 0?De(h.connectTime):"—",m=h.disconnectTime!==void 0?De(h.disconnectTime):"—",O=Object.entries(h.status).filter(([I])=>!["total","lastUpdateTime"].includes(I)).map(([I,x])=>{const $=`${x.responsesAcknowledged??0} / ${x.responsesSent??0}`;return{type:I.toUpperCase(),ratio:$,responsesSent:x.responsesSent??0,responsesAcknowledged:x.responsesAcknowledged??0,responsesRejected:x.responsesRejected??0}});return{subscription:h,formattedConnectDate:p,formattedDisconnectDate:m,statuses:O}})}),c=S(()=>{const{status:r}=xe(l.dataPlaneOverview.dataplane,l.dataPlaneOverview.dataplaneInsight);return r}),k=S(()=>{const r=Be(l.dataPlaneOverview.dataplaneInsight);return r!==null?Object.entries(r).map(([y,h])=>({name:y,version:h})):[]}),g=S(()=>{var x;const r=((x=l.dataPlaneOverview.dataplaneInsight)==null?void 0:x.subscriptions)??[];if(r.length===0)return[];const y=r[r.length-1];if(!y.version)return[];const h=[],p=y.version.envoy,m=y.version.kumaDp;if(!(p.kumaDpCompatible!==void 0?p.kumaDpCompatible:!0)){const $=`Envoy ${p.version} is not supported by Kuma DP ${m.version}.`;h.push($)}if(!(m.kumaCpCompatible!==void 0?m.kumaCpCompatible:!0)){const $=`Kuma DP ${m.version} is not supported by this Kuma control plane.`;h.push($)}return h});return(r,y)=>{const h=He("router-link");return u(),N(d(nt),null,{body:C(()=>[_("div",Vt,[_("section",null,[_("h3",Nt,[_("span",null,[o(`
              DPP:

              `),D(h,{to:{name:"data-plane-detail-view",params:{mesh:e.dataPlaneOverview.mesh,dataPlane:e.dataPlaneOverview.name}}},{default:C(()=>[o(P(e.dataPlaneOverview.name),1)]),_:1},8,["to"])]),o(),D(ut,{status:d(c)},null,8,["status"])]),o(),D(re,{class:"mt-4"},{default:C(()=>[D(K,{term:"Mesh"},{default:C(()=>[o(P(e.dataPlaneOverview.mesh),1)]),_:1}),o(),d(T)!==null?(u(),N(K,{key:0,term:"Tags"},{default:C(()=>[D(dt,{tags:d(T)},null,8,["tags"])]),_:1})):V("",!0),o(),d(k).length>0?(u(),N(K,{key:1,term:"Dependencies"},{default:C(()=>[_("ul",null,[(u(!0),w(F,null,M(d(k),(p,m)=>(u(),w("li",{key:m},P(p.name)+": "+P(p.version),1))),128))]),o(),d(g).length>0?(u(),w(F,{key:0},[_("h5",Ft,[o(`
                  Warnings

                  `),D(d(X),{class:"ml-1",icon:"warning",color:"var(--black-500)","secondary-color":"var(--yellow-300)",size:"20"})]),o(),(u(!0),w(F,null,M(d(g),(p,m)=>(u(),w("p",{key:m},P(p),1))),128))],64)):V("",!0)]),_:1})):V("",!0)]),_:1})]),o(),d(i).length>0?(u(),w("section",Kt,[Mt,o(),_("div",qt,[(u(!0),w(F,null,M(d(i),(p,m)=>(u(),w("div",{key:m},[D(re,null,{default:C(()=>[D(K,{term:"Connect time","data-testid":`data-plane-connect-time-${m}`},{default:C(()=>[o(P(p.formattedConnectDate),1)]),_:2},1032,["data-testid"]),o(),D(K,{term:"Disconnect time","data-testid":`data-plane-disconnect-time-${m}`},{default:C(()=>[o(P(p.formattedDisconnectDate),1)]),_:2},1032,["data-testid"]),o(),D(K,{term:"CP instance ID"},{default:C(()=>[o(P(p.subscription.controlPlaneInstanceId),1)]),_:2},1024)]),_:2},1024),o(),p.statuses.length>0?(u(),w("details",Rt,[jt,o(),D(re,null,{default:C(()=>[(u(!0),w(F,null,M(p.statuses,(O,I)=>(u(),N(K,{key:`${m}-${I}`,term:O.type,"data-testid":`data-plane-subscription-status-${m}-${I}`},{default:C(()=>[o(P(O.ratio),1)]),_:2},1032,["term","data-testid"]))),128))]),_:2},1024)])):V("",!0)]))),128))])])):V("",!0),o(),_("section",zt,[D(ct,{id:"code-block-data-plane-summary",content:d(t),"code-max-height":"250px"},null,8,["content"])])])]),_:1})}}});const Ht=pe(Bt,[["__scopeId","data-v-3fa16dcc"]]),Qt=["protocol","service","zone"];function Gt(e){const l=new Map;for(const[t,T]of e){const i=Qt.includes(t),c=i?"tag":t;l.has(c)||l.set(c,[]);const k=l.get(c);let g;c==="tag"?g=(i?`kuma.io/${t}:${T}`:T).replace(/\s+/g,""):g=T,k.push(g.trim())}return l}const Yt=e=>(de("data-v-7f65c55f"),e=e(),ce(),e),Jt={key:0},Zt=Yt(()=>_("label",{for:"data-planes-type-filter",class:"mr-2"},`
              Type:
            `,-1)),Wt=["value"],Xt=["for"],ea=["id","checked","onChange"],ta=ue({__name:"DataPlaneList",props:{dataPlaneOverviews:{type:Array,required:!0},isLoading:{type:Boolean,required:!1,default:!1},error:{type:[Error,null],required:!1,default:null},nextUrl:{type:String,required:!1,default:null},pageOffset:{type:Number,required:!1,default:0},selectedDppName:{type:String,required:!1,default:null},isGatewayView:{type:Boolean,required:!1,default:!1},gatewayType:{type:String,required:!1,default:"true"},dppFilterFields:{type:Object,required:!0}},emits:["load-data"],setup(e,{emit:l}){const t=e,T={true:"All",builtin:"Builtin",delegated:"Delegated"},i={title:"No Data",message:"There are no data plane proxies present."},c=Qe(),k=pt(),g=E(Ie),r=E({headers:[],data:[]}),y=E(R.get("filterQuery")??""),h=E(t.gatewayType),p=E({}),m=E(null),O=S(()=>k.getters["config/getMulticlusterStatus"]),I=S(()=>({name:k.getters["config/getEnvironment"]==="universal"?"universal-dataplane":"kubernetes-dataplane"})),x=S(()=>"tag"in t.dppFilterFields?'tag: "kuma.io/protocol: http"':"name"in t.dppFilterFields?"name: cluster":"field: value"),$=S(()=>{let a=gt(O.value,g.value);return t.isGatewayView?a=a.filter(v=>v.key!=="protocol"):a=a.filter(v=>v.key!=="type"),{data:r.value.data,headers:a}}),ee=S(()=>vt.filter(a=>t.isGatewayView?a.tableHeaderKey!=="protocol":a.tableHeaderKey!=="type").filter(a=>O.value?!0:a.tableHeaderKey!=="zone").map(a=>{const v=g.value.includes(a.tableHeaderKey);return{...a,isChecked:v}}));Q(h,function(){j(0)}),Q(p,function(){j(0)}),Q(()=>t.dataPlaneOverviews,function(){z()});function G(){const a=Ce.get("dpVisibleTableHeaderKeys");Array.isArray(a)&&(g.value=a),z()}G();function te(a){j(a)}function j(a){const v={...p.value};"gateway"in v||(v.gateway=h.value),l("load-data",a,v)}function Y(a){a.stopPropagation()}function ae(a,v){const n=a.target,s=g.value.findIndex(f=>f===v);n.checked&&s===-1?g.value.push(v):!n.checked&&s>-1&&g.value.splice(s,1),Ce.set("dpVisibleTableHeaderKeys",Array.from(new Set(g.value)))}function ne(){at.logger.info(ft.CREATE_DATA_PLANE_PROXY_CLICKED)}async function z(){try{Array.isArray(t.dataPlaneOverviews)&&t.dataPlaneOverviews.length>0?(B(t.selectedDppName??t.dataPlaneOverviews[0].name),r.value.data=await Promise.all(t.dataPlaneOverviews.map(a=>se(a)))):(B(null),r.value.data=[])}catch(a){console.error(a)}}function B(a){a&&t.dataPlaneOverviews.length>0?(m.value=t.dataPlaneOverviews.find(v=>v.name===a)??t.dataPlaneOverviews[0],R.set(t.isGatewayView?"gateway":"dpp",m.value.name)):(m.value=null,R.set(t.isGatewayView?"gateway":"dpp",null))}async function se(a){var ve,ge,ye,he,be;const v=a.mesh,n=a.name,s=((ve=a.dataplane.networking.gateway)==null?void 0:ve.type)||"STANDARD",f={name:s==="STANDARD"?"data-plane-detail-view":"gateway-detail-view",params:{mesh:v,dataPlane:n}},U={name:"mesh-detail-view",params:{mesh:v}},J=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],Z=Ee(a.dataplane).filter(b=>J.includes(b.label)),le=(ge=Z.find(b=>b.label==="kuma.io/service"))==null?void 0:ge.value,Le=(ye=Z.find(b=>b.label==="kuma.io/protocol"))==null?void 0:ye.value,oe=(he=Z.find(b=>b.label==="kuma.io/zone"))==null?void 0:he.value;let fe;le!==void 0&&(fe={name:"service-detail-view",params:{mesh:v,service:le}});let me;oe!==void 0&&(me={name:"zone-list-view",query:{ns:oe}});const{status:Ve}=xe(a.dataplane,a.dataplaneInsight),Ne=((be=a.dataplaneInsight)==null?void 0:be.subscriptions)??[],Fe={totalUpdates:0,totalRejectedUpdates:0,dpVersion:null,envoyVersion:null,selectedTime:NaN,selectedUpdateTime:NaN,version:null},A=Ne.reduce((b,L)=>{var _e,ke;if(L.connectTime){const we=Date.parse(L.connectTime);(!b.selectedTime||we>b.selectedTime)&&(b.selectedTime=we)}const ie=Date.parse(L.status.lastUpdateTime);return ie&&(!b.selectedUpdateTime||ie>b.selectedUpdateTime)&&(b.selectedUpdateTime=ie),{totalUpdates:b.totalUpdates+parseInt(L.status.total.responsesSent??"0",10),totalRejectedUpdates:b.totalRejectedUpdates+parseInt(L.status.total.responsesRejected??"0",10),dpVersion:((_e=L.version)==null?void 0:_e.kumaDp.version)||b.dpVersion,envoyVersion:((ke=L.version)==null?void 0:ke.envoy.version)||b.envoyVersion,selectedTime:b.selectedTime,selectedUpdateTime:b.selectedUpdateTime,version:L.version||b.version}},Fe),q={name:n,nameRoute:f,mesh:v,meshRoute:U,type:s,zone:oe??"—",zoneRoute:me,service:le??"—",serviceInsightRoute:fe,protocol:Le??"—",status:Ve,totalUpdates:A.totalUpdates,totalRejectedUpdates:A.totalRejectedUpdates,dpVersion:A.dpVersion??"—",envoyVersion:A.envoyVersion??"—",warnings:[],unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:A.selectedUpdateTime?Pe(new Date(A.selectedUpdateTime).toUTCString()):"—",lastConnected:A.selectedTime?Pe(new Date(A.selectedTime).toUTCString()):"—",overview:a};if(A.version){const{kind:b}=Ge(A.version);switch(b!==Ye&&q.warnings.push(b),b){case Ze:q.unsupportedEnvoyVersion=!0;break;case Je:q.unsupportedKumaDPVersion=!0;break}}return O.value&&A.dpVersion&&Z.find(L=>L.label===We)&&typeof A.version.kumaDp.kumaCpCompatible=="boolean"&&!A.version.kumaDp.kumaCpCompatible&&(q.warnings.push(Xe),q.kumaDpAndKumaCpMismatch=!0),q}function H({fields:a,query:v}){const n=R.get("filterFields"),s=n!==null?JSON.parse(n):{},f=JSON.stringify(s),U=Object.fromEntries(Gt(a)),J=JSON.stringify(U);R.set("filterQuery",v||null),R.set("filterFields",J),f!==J&&(p.value=U)}return(a,v)=>(u(),N(ot,null,{content:C(()=>{var n;return[D(it,{"selected-entity-name":(n=m.value)==null?void 0:n.name,"page-size":d(tt),"is-loading":t.isLoading,error:e.error,"empty-state":i,"table-data":d($),"table-data-is-empty":d($).data.length===0,"show-details":"",next:t.nextUrl!==null,"page-offset":t.pageOffset,onTableAction:v[1]||(v[1]=s=>B(s.name)),onLoadData:te},{additionalControls:C(()=>[D(Lt,{id:"data-plane-proxy-filter",class:"data-plane-proxy-filter",placeholder:d(x),query:y.value,fields:t.dppFilterFields,onFieldsChange:H},null,8,["placeholder","query","fields"]),o(),t.isGatewayView?(u(),w("div",Jt,[Zt,o(),Ae(_("select",{id:"data-planes-type-filter","onUpdate:modelValue":v[0]||(v[0]=s=>h.value=s),"data-testid":"data-planes-type-filter"},[(u(!0),w(F,null,M(d(T),(s,f)=>(u(),w("option",{key:f,value:f},P(s),9,Wt))),128))],512),[[et,h.value]])])):V("",!0),o(),D(d(st),{label:"Columns",icon:"cogwheel","button-appearance":"outline"},{items:C(()=>[_("div",{onClick:Y},[(u(!0),w(F,null,M(d(ee),(s,f)=>(u(),N(d(lt),{key:f,class:"table-header-selector-item",item:s},{default:C(()=>[_("label",{for:`data-plane-table-header-checkbox-${f}`,class:"k-checkbox table-header-selector-item-checkbox"},[_("input",{id:`data-plane-table-header-checkbox-${f}`,checked:s.isChecked,type:"checkbox",class:"k-input",onChange:U=>ae(U,s.tableHeaderKey)},null,40,ea),o(" "+P(s.label),1)],8,Xt)]),_:2},1032,["item"]))),128))])]),_:1}),o(),D(d(Se),{appearance:"creation",to:d(I),icon:"plus","data-testid":"data-plane-create-data-plane-button",onClick:ne},{default:C(()=>[o(`
            Create data plane proxy
          `)]),_:1},8,["to"]),o(),d(c).query.ns?(u(),N(d(Se),{key:1,appearance:"primary",icon:"arrowLeft",to:{name:d(c).name},"data-testid":"data-plane-ns-back-button"},{default:C(()=>[o(`
            View All
          `)]),_:1},8,["to"])):V("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","next","page-offset"])]}),sidebar:C(()=>[m.value!==null?(u(),N(Ht,{key:0,"data-plane-overview":m.value},null,8,["data-plane-overview"])):(u(),N(rt,{key:1}))]),_:1}))}});const va=pe(ta,[["__scopeId","data-v-7f65c55f"]]);export{va as D};
