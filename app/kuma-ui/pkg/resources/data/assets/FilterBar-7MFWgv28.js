var se=Object.defineProperty;var le=(n,i,e)=>i in n?se(n,i,{enumerable:!0,configurable:!0,writable:!0,value:e}):n[i]=e;var N=(n,i,e)=>(le(n,typeof i!="symbol"?i+"":i,e),e);import{d as K,l as ie,a as q,o,b as w,W as oe,w as m,r as O,t as d,f as r,e as S,q as c,F as x,V as re,c as g,H as Q,p as I,m as _,T as ue,K as L,U as ce,_ as R,ac as de,C as T,M as A,a6 as E,ao as fe,ap as pe,aq as me,n as j,ar as ge,as as ye,at as he,au as ve,x as be,y as ke}from"./index-eDf0gHDD.js";import{A as _e}from"./AppCollection-ReyhGgL1.js";import{S as Se}from"./StatusBadge-HXpDGqsA.js";const xe={key:0},Ce={key:1},Te=K({__name:"DataPlaneList",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{},isSelectedRow:{type:[Function,null],default:null},summaryRouteName:{},isGlobalMode:{type:Boolean},canUseGatewaysUi:{type:Boolean,default:!1}},emits:["change"],setup(n,{emit:i}){const{t:e}=ie(),f=n,h=i;return(l,s)=>{const b=q("RouterLink"),v=q("KTruncate"),k=q("KTooltip");return o(),w(_e,{class:"data-plane-collection","empty-state-message":c(e)("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":c(e)("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":c(e)("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Type",key:"type"},{label:"Services",key:"services"},...f.isGlobalMode?[{label:"Zone",key:"zone"}]:[],{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],"page-number":f.pageNumber,"page-size":f.pageSize,total:f.total,items:f.items,error:f.error,"is-selected-row":f.isSelectedRow,onChange:s[0]||(s[0]=t=>h("change",t))},oe({name:m(({row:t})=>[S(b,{class:"name-link",title:t.name,to:{name:f.summaryRouteName,params:{mesh:t.mesh,dataPlane:t.name},query:{page:f.pageNumber,size:f.pageSize}}},{default:m(()=>[r(d(t.name),1)]),_:2},1032,["title","to"])]),type:m(({row:t})=>[r(d(c(e)(`data-planes.type.${t.dataplaneType}`)),1)]),services:m(({row:t})=>[t.services.length>0?(o(),w(v,{key:0,width:"auto"},{default:m(()=>[(o(!0),g(x,null,Q(t.services,(p,C)=>(o(),g("div",{key:C},[S(re,{text:p},{default:m(()=>[t.dataplaneType==="standard"?(o(),w(b,{key:0,to:{name:"service-detail-view",params:{service:p}}},{default:m(()=>[r(d(p),1)]),_:2},1032,["to"])):t.dataplaneType==="delegated"&&f.canUseGatewaysUi?(o(),w(b,{key:1,to:{name:"delegated-gateway-detail-view",params:{service:p}}},{default:m(()=>[r(d(p),1)]),_:2},1032,["to"])):(o(),g(x,{key:2},[r(d(p),1)],64))]),_:2},1032,["text"])]))),128))]),_:2},1024)):(o(),g(x,{key:1},[r(d(c(e)("common.collection.none")),1)],64))]),zone:m(({row:t})=>[t.zone?(o(),w(b,{key:0,to:{name:"zone-cp-detail-view",params:{zone:t.zone}}},{default:m(()=>[r(d(t.zone),1)]),_:2},1032,["to"])):(o(),g(x,{key:1},[r(d(c(e)("common.collection.none")),1)],64))]),certificate:m(({row:t})=>{var p;return[(p=t.dataplaneInsight.mTLS)!=null&&p.certificateExpirationTime?(o(),g(x,{key:0},[r(d(c(e)("common.formats.datetime",{value:Date.parse(t.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(o(),g(x,{key:1},[r(d(c(e)("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:m(({row:t})=>[S(Se,{status:t.status},null,8,["status"])]),warnings:m(({row:t})=>[t.isCertExpired||t.warnings.length>0?(o(),w(k,{key:0},{content:m(()=>[_("ul",null,[t.warnings.length>0?(o(),g("li",xe,d(c(e)("data-planes.components.data-plane-list.version_mismatch")),1)):I("",!0),r(),t.isCertExpired?(o(),g("li",Ce,d(c(e)("data-planes.components.data-plane-list.cert_expired")),1)):I("",!0)])]),default:m(()=>[r(),S(ue,{class:"mr-1",size:c(L),"hide-title":""},null,8,["size"])]),_:2},1024)):(o(),g(x,{key:1},[r(d(c(e)("common.collection.none")),1)],64))]),details:m(({row:t})=>[S(b,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:t.name}}},{default:m(()=>[r(d(c(e)("common.collection.details_link"))+" ",1),S(c(ce),{display:"inline-block",decorative:"",size:c(L)},null,8,["size"])]),_:2},1032,["to"])]),_:2},[l.$slots.toolbar?{name:"toolbar",fn:m(()=>[O(l.$slots,"toolbar",{},void 0,!0)]),key:"0"}:void 0]),1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error","is-selected-row"])}}}),We=R(Te,[["__scopeId","data-v-84083e88"]]);function we(n,i,e){return Math.max(i,Math.min(n,e))}const ze=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class Ie{constructor(i,e){N(this,"commands");N(this,"keyMap");N(this,"boundTriggerShortcuts");this.commands=e,this.keyMap=Object.fromEntries(Object.entries(i).map(([f,h])=>[f.toLowerCase(),h])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(i){Le(i,this.keyMap,this.commands)}}function Le(n,i,e){const f=Fe(n.code),h=[n.ctrlKey?"ctrl":"",n.shiftKey?"shift":"",n.altKey?"alt":"",f].filter(b=>b!=="").join("+"),l=i[h];if(!l)return;const s=e[l];s.isAllowedContext&&!s.isAllowedContext(n)||(s.shouldPreventDefaultAction&&n.preventDefault(),!(s.isDisabled&&s.isDisabled())&&s.trigger(n))}function Fe(n){return ze.includes(n)?"":n.replace(/^Key/,"").toLowerCase()}function Ne(n,i){const e=" "+n,f=e.matchAll(/ ([-\s\w]+):\s*/g),h=[];for(const l of Array.from(f)){if(l.index===void 0)continue;const s=Ae(l[1]);if(i.length>0&&!i.includes(s))throw new Error(`Unknown field “${s}”. Known fields: ${i.join(", ")}`);const b=l.index+l[0].length,v=e.substring(b);let k;if(/^\s*["']/.test(v)){const p=v.match(/['"](.*?)['"]/);if(p!==null)k=p[1];else throw new Error(`Quote mismatch for field “${s}”.`)}else{const p=v.indexOf(" "),C=p===-1?v.length:p;k=v.substring(0,C)}k!==""&&h.push([s,k])}return h}function Ae(n){return n.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(i,e)=>e===0?i:i.substring(1).toUpperCase())}const U=n=>(be("data-v-f8c4e95f"),n=n(),ke(),n),Pe=U(()=>_("span",{class:"visually-hidden"},"Focus filter",-1)),qe={class:"k-filter-icon"},Ee=["for"],Be=["id","placeholder"],Me={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},De={class:"k-suggestion-list"},$e={key:0,class:"k-filter-bar-error"},je={key:0},Ke=["title","data-filter-field"],Oe={class:"visually-hidden"},Qe=U(()=>_("span",{class:"visually-hidden"},"Clear query",-1)),Re=K({__name:"FilterBar",props:{id:{type:String,required:!1,default:()=>de("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(n,{emit:i}){const e=n,f=i,h=T(null),l=T(null),s=T(e.query),b=T([]),v=T(null),k=T(!1),t=T(-1),p=A(()=>Object.keys(e.fields)),C=A(()=>Object.entries(e.fields).slice(0,5).map(([a,u])=>({fieldName:a,...u}))),B=A(()=>p.value.length>0?`Filter by ${p.value.join(", ")}`:"Filter"),V=A(()=>e.placeholder??B.value);E(()=>b.value,function(a,u){ne(a,u)||(v.value=null,f("fields-change",{fields:a,query:s.value}))}),E(()=>e.query,()=>{s.value=e.query,z(s.value)},{immediate:!0}),E(()=>s.value,function(){s.value===""&&(v.value=null)});const H={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},G={submitQuery:{trigger:M,isAllowedContext(a){return l.value!==null&&a.composedPath().includes(l.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:Z,isAllowedContext(a){return l.value!==null&&a.composedPath().includes(l.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:Y,isAllowedContext(a){return l.value!==null&&a.composedPath().includes(l.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:P,isAllowedContext(a){return h.value!==null&&a.composedPath().includes(h.value)}}};function W(){const a=new Ie(H,G);he(function(){a.registerListener()}),ve(function(){a.unRegisterListener()}),z(s.value)}W();function J(a){const u=a.target;z(u.value)}function M(){if(l.value instanceof HTMLInputElement)if(t.value===-1)z(l.value.value),k.value=!1;else{const a=C.value[t.value].fieldName;a&&$(l.value,a)}}function Z(){D(1)}function Y(){D(-1)}function D(a){t.value=we(t.value+a,-1,C.value.length-1)}function X(){l.value instanceof HTMLInputElement&&l.value.focus()}function ee(a){const y=a.currentTarget.getAttribute("data-filter-field");y&&l.value instanceof HTMLInputElement&&$(l.value,y)}function $(a,u){const y=s.value===""||s.value.endsWith(" ")?"":" ";s.value+=y+u+":",a.focus(),t.value=-1}function te(){s.value="",l.value instanceof HTMLInputElement&&(l.value.value="",l.value.focus(),z(""))}function ae(a){a.relatedTarget===null&&P(),h.value instanceof HTMLElement&&a.relatedTarget instanceof Node&&!h.value.contains(a.relatedTarget)&&P()}function P(){k.value=!1}function z(a){v.value=null;try{const u=Ne(a,p.value);u.sort((y,F)=>y[0].localeCompare(F[0])),b.value=u}catch(u){if(u instanceof Error)v.value=u,k.value=!0;else throw u}}function ne(a,u){return JSON.stringify(a)===JSON.stringify(u)}return(a,u)=>(o(),g("div",{ref_key:"filterBar",ref:h,class:"k-filter-bar","data-testid":"k-filter-bar"},[_("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:X},[Pe,r(),_("span",qe,[S(c(fe),{decorative:"","data-testid":"k-filter-bar-filter-icon","hide-title":"",size:c(L)},null,8,["size"])])]),r(),_("label",{for:`${e.id}-filter-bar-input`,class:"visually-hidden"},[O(a.$slots,"default",{},()=>[r(d(B.value),1)],!0)],8,Ee),r(),pe(_("input",{id:`${e.id}-filter-bar-input`,ref_key:"filterInput",ref:l,"onUpdate:modelValue":u[0]||(u[0]=y=>s.value=y),class:"k-filter-bar-input",type:"text",placeholder:V.value,"data-testid":"k-filter-bar-filter-input",onFocus:u[1]||(u[1]=y=>k.value=!0),onBlur:ae,onChange:J},null,40,Be),[[me,s.value]]),r(),k.value?(o(),g("div",Me,[_("div",De,[v.value!==null?(o(),g("p",$e,d(v.value.message),1)):(o(),g("button",{key:1,class:j(["k-submit-query-button",{"k-submit-query-button-is-selected":t.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:M},`
          Submit `+d(s.value),3)),r(),(o(!0),g(x,null,Q(C.value,(y,F)=>(o(),g("div",{key:`${e.id}-${F}`,class:j(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":t.value===F}])},[_("b",null,d(y.fieldName),1),y.description!==""?(o(),g("span",je,": "+d(y.description),1)):I("",!0),r(),_("button",{class:"k-apply-suggestion-button",title:`Add ${y.fieldName}:`,type:"button","data-filter-field":y.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:ee},[_("span",Oe,"Add "+d(y.fieldName)+":",1),r(),S(c(ge),{decorative:"","hide-title":"",size:c(L)},null,8,["size"])],8,Ke)],2))),128))])])):I("",!0),r(),s.value!==""?(o(),g("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:te},[Qe,r(),S(c(ye),{decorative:"","hide-title":"",size:c(L)},null,8,["size"])])):I("",!0)],512))}}),Je=R(Re,[["__scopeId","data-v-f8c4e95f"]]);export{We as D,Je as F};
