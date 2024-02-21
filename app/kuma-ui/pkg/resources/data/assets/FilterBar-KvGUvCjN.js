var se=Object.defineProperty;var ie=(n,o,e)=>o in n?se(n,o,{enumerable:!0,configurable:!0,writable:!0,value:e}):n[o]=e;var F=(n,o,e)=>(ie(n,typeof o!="symbol"?o+"":o,e),e);import{d as K,l as oe,a as E,o as c,b as N,a5 as le,w as p,r as O,t as f,f as r,e as S,q as u,F as T,W as re,c as m,H as Q,p as w,m as _,T as ue,K as z,U as ce,_ as R,ad as de,C,M as A,a7 as $,ap as fe,aq as pe,ar as me,n as j,as as ge,at as ve,au as he,av as ye,x as be,y as ke}from"./index-KQusT94Q.js";import{A as _e}from"./AppCollection-WlSbbR7B.js";import{S as Se}from"./StatusBadge-e0YwD099.js";const xe={key:0},Ce={key:1},Te=K({__name:"DataPlaneList",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{},isSelectedRow:{type:[Function,null],default:null},summaryRouteName:{},isGlobalMode:{type:Boolean}},emits:["change"],setup(n,{emit:o}){const{t:e}=oe(),d=n,h=o;return(s,i)=>{const k=E("RouterLink"),y=E("KTruncate"),b=E("KTooltip");return c(),N(_e,{class:"data-plane-collection","empty-state-message":u(e)("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":u(e)("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":u(e)("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Type",key:"type"},{label:"Services",key:"services"},...d.isGlobalMode?[{label:"Zone",key:"zone"}]:[],{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],"page-number":d.pageNumber,"page-size":d.pageSize,total:d.total,items:d.items,error:d.error,"is-selected-row":d.isSelectedRow,onChange:i[0]||(i[0]=t=>h("change",t))},le({name:p(({row:t})=>[S(k,{class:"name-link",title:t.name,to:{name:d.summaryRouteName,params:{mesh:t.mesh,dataPlane:t.name},query:{page:d.pageNumber,size:d.pageSize}}},{default:p(()=>[r(f(t.name),1)]),_:2},1032,["title","to"])]),type:p(({row:t})=>[r(f(u(e)(`data-planes.type.${t.dataplaneType}`)),1)]),services:p(({row:t})=>[t.services.length>0?(c(),N(y,{key:0,width:"auto"},{default:p(()=>[(c(!0),m(T,null,Q(t.services,(g,x)=>(c(),m("div",{key:x},[S(re,{text:g},{default:p(()=>[S(k,{to:{name:"service-detail-view",params:{service:g}}},{default:p(()=>[r(f(g),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]))),128))]),_:2},1024)):(c(),m(T,{key:1},[r(f(u(e)("common.collection.none")),1)],64))]),zone:p(({row:t})=>[t.zone?(c(),N(k,{key:0,to:{name:"zone-cp-detail-view",params:{zone:t.zone}}},{default:p(()=>[r(f(t.zone),1)]),_:2},1032,["to"])):(c(),m(T,{key:1},[r(f(u(e)("common.collection.none")),1)],64))]),certificate:p(({row:t})=>{var g;return[(g=t.dataplaneInsight.mTLS)!=null&&g.certificateExpirationTime?(c(),m(T,{key:0},[r(f(u(e)("common.formats.datetime",{value:Date.parse(t.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(c(),m(T,{key:1},[r(f(u(e)("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:p(({row:t})=>[S(Se,{status:t.status},null,8,["status"])]),warnings:p(({row:t})=>[t.isCertExpired||t.warnings.length>0?(c(),N(b,{key:0},{content:p(()=>[_("ul",null,[t.warnings.length>0?(c(),m("li",xe,f(u(e)("data-planes.components.data-plane-list.version_mismatch")),1)):w("",!0),r(),t.isCertExpired?(c(),m("li",Ce,f(u(e)("data-planes.components.data-plane-list.cert_expired")),1)):w("",!0)])]),default:p(()=>[r(),S(ue,{class:"mr-1",size:u(z),"hide-title":""},null,8,["size"])]),_:2},1024)):(c(),m(T,{key:1},[r(f(u(e)("common.collection.none")),1)],64))]),details:p(({row:t})=>[S(k,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:t.name}}},{default:p(()=>[r(f(u(e)("common.collection.details_link"))+" ",1),S(u(ce),{display:"inline-block",decorative:"",size:u(z)},null,8,["size"])]),_:2},1032,["to"])]),_:2},[s.$slots.toolbar?{name:"toolbar",fn:p(()=>[O(s.$slots,"toolbar",{},void 0,!0)]),key:"0"}:void 0]),1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error","is-selected-row"])}}}),Ge=R(Te,[["__scopeId","data-v-fe351bd8"]]);function we(n,o,e){return Math.max(o,Math.min(n,e))}const ze=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class Ie{constructor(o,e){F(this,"commands");F(this,"keyMap");F(this,"boundTriggerShortcuts");this.commands=e,this.keyMap=Object.fromEntries(Object.entries(o).map(([d,h])=>[d.toLowerCase(),h])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(o){Le(o,this.keyMap,this.commands)}}function Le(n,o,e){const d=Fe(n.code),h=[n.ctrlKey?"ctrl":"",n.shiftKey?"shift":"",n.altKey?"alt":"",d].filter(k=>k!=="").join("+"),s=o[h];if(!s)return;const i=e[s];i.isAllowedContext&&!i.isAllowedContext(n)||(i.shouldPreventDefaultAction&&n.preventDefault(),!(i.isDisabled&&i.isDisabled())&&i.trigger(n))}function Fe(n){return ze.includes(n)?"":n.replace(/^Key/,"").toLowerCase()}function Ne(n,o){const e=" "+n,d=e.matchAll(/ ([-\s\w]+):\s*/g),h=[];for(const s of Array.from(d)){if(s.index===void 0)continue;const i=Ae(s[1]);if(o.length>0&&!o.includes(i))throw new Error(`Unknown field “${i}”. Known fields: ${o.join(", ")}`);const k=s.index+s[0].length,y=e.substring(k);let b;if(/^\s*["']/.test(y)){const g=y.match(/['"](.*?)['"]/);if(g!==null)b=g[1];else throw new Error(`Quote mismatch for field “${i}”.`)}else{const g=y.indexOf(" "),x=g===-1?y.length:g;b=y.substring(0,x)}b!==""&&h.push([i,b])}return h}function Ae(n){return n.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(o,e)=>e===0?o:o.substring(1).toUpperCase())}const V=n=>(be("data-v-d51e0350"),n=n(),ke(),n),Pe=V(()=>_("span",{class:"visually-hidden"},"Focus filter",-1)),Ee={class:"k-filter-icon"},Me=["for"],qe=["id","placeholder"],Be={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},De={class:"k-suggestion-list"},$e={key:0,class:"k-filter-bar-error"},je={key:0},Ke=["title","data-filter-field"],Oe={class:"visually-hidden"},Qe=V(()=>_("span",{class:"visually-hidden"},"Clear query",-1)),Re=K({__name:"FilterBar",props:{id:{type:String,required:!1,default:()=>de("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(n,{emit:o}){const e=n,d=o,h=C(null),s=C(null),i=C(e.query),k=C([]),y=C(null),b=C(!1),t=C(-1),g=A(()=>Object.keys(e.fields)),x=A(()=>Object.entries(e.fields).slice(0,5).map(([a,l])=>({fieldName:a,...l}))),M=A(()=>g.value.length>0?`Filter by ${g.value.join(", ")}`:"Filter"),U=A(()=>e.placeholder??M.value);$(()=>k.value,function(a,l){ne(a,l)||(y.value=null,d("fields-change",{fields:a,query:i.value}))}),$(()=>i.value,function(){i.value===""&&(y.value=null),b.value=!0});const H={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},W={submitQuery:{trigger:q,isAllowedContext(a){return s.value!==null&&a.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:Z,isAllowedContext(a){return s.value!==null&&a.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:Y,isAllowedContext(a){return s.value!==null&&a.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:P,isAllowedContext(a){return h.value!==null&&a.composedPath().includes(h.value)}}};function G(){const a=new Ie(H,W);he(function(){a.registerListener()}),ye(function(){a.unRegisterListener()}),I(i.value)}G();function J(a){const l=a.target;I(l.value)}function q(){if(s.value instanceof HTMLInputElement)if(t.value===-1)I(s.value.value),b.value=!1;else{const a=x.value[t.value].fieldName;a&&D(s.value,a)}}function Z(){B(1)}function Y(){B(-1)}function B(a){t.value=we(t.value+a,-1,x.value.length-1)}function X(){s.value instanceof HTMLInputElement&&s.value.focus()}function ee(a){const v=a.currentTarget.getAttribute("data-filter-field");v&&s.value instanceof HTMLInputElement&&D(s.value,v)}function D(a,l){const v=i.value===""||i.value.endsWith(" ")?"":" ";i.value+=v+l+":",a.focus(),t.value=-1}function te(){i.value="",s.value instanceof HTMLInputElement&&(s.value.value="",s.value.focus(),I(""))}function ae(a){a.relatedTarget===null&&P(),h.value instanceof HTMLElement&&a.relatedTarget instanceof Node&&!h.value.contains(a.relatedTarget)&&P()}function P(){b.value=!1}function I(a){y.value=null;try{const l=Ne(a,g.value);l.sort((v,L)=>v[0].localeCompare(L[0])),k.value=l}catch(l){if(l instanceof Error)y.value=l,b.value=!0;else throw l}}function ne(a,l){return JSON.stringify(a)===JSON.stringify(l)}return(a,l)=>(c(),m("div",{ref_key:"filterBar",ref:h,class:"k-filter-bar","data-testid":"k-filter-bar"},[_("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:X},[Pe,r(),_("span",Ee,[S(u(fe),{decorative:"","data-testid":"k-filter-bar-filter-icon","hide-title":"",size:u(z)},null,8,["size"])])]),r(),_("label",{for:`${e.id}-filter-bar-input`,class:"visually-hidden"},[O(a.$slots,"default",{},()=>[r(f(M.value),1)],!0)],8,Me),r(),pe(_("input",{id:`${e.id}-filter-bar-input`,ref_key:"filterInput",ref:s,"onUpdate:modelValue":l[0]||(l[0]=v=>i.value=v),class:"k-filter-bar-input",type:"text",placeholder:U.value,"data-testid":"k-filter-bar-filter-input",onFocus:l[1]||(l[1]=v=>b.value=!0),onBlur:ae,onChange:J},null,40,qe),[[me,i.value]]),r(),b.value?(c(),m("div",Be,[_("div",De,[y.value!==null?(c(),m("p",$e,f(y.value.message),1)):(c(),m("button",{key:1,class:j(["k-submit-query-button",{"k-submit-query-button-is-selected":t.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:q},`
          Submit `+f(i.value),3)),r(),(c(!0),m(T,null,Q(x.value,(v,L)=>(c(),m("div",{key:`${e.id}-${L}`,class:j(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":t.value===L}])},[_("b",null,f(v.fieldName),1),v.description!==""?(c(),m("span",je,": "+f(v.description),1)):w("",!0),r(),_("button",{class:"k-apply-suggestion-button",title:`Add ${v.fieldName}:`,type:"button","data-filter-field":v.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:ee},[_("span",Oe,"Add "+f(v.fieldName)+":",1),r(),S(u(ge),{decorative:"","hide-title":"",size:u(z)},null,8,["size"])],8,Ke)],2))),128))])])):w("",!0),r(),i.value!==""?(c(),m("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:te},[Qe,r(),S(u(ve),{decorative:"","hide-title":"",size:u(z)},null,8,["size"])])):w("",!0)],512))}}),Je=R(Re,[["__scopeId","data-v-d51e0350"]]);export{Ge as D,Je as F};
