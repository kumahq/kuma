var le=Object.defineProperty;var re=(n,i,a)=>i in n?le(n,i,{enumerable:!0,configurable:!0,writable:!0,value:a}):n[i]=a;var K=(n,i,a)=>(re(n,typeof i!="symbol"?i+"":i,a),a);import{d as oe,$ as ue,I as ce,r as de,o as f,g as z,w as h,P as ne,h as C,l as m,C as k,j as _,F as E,S as pe,i as c,a5 as me,V as fe,K as R,am as ge,U as ve,a2 as $,Z as ye,an as he,Y as be,q as se,t as A,f as q,ao as ee,m as S,ap as _e,aq as ke,ar as Te,x as te,D as Se,k as W,v as Ce,Q as we,z as xe,B as Ae}from"./index-65a641bf.js";import{K as Ie}from"./constants-b84da159.js";import{d as Ue,a as De,c as Ne,C as Le,e as ze}from"./dataplane-a974028d.js";const Ee=oe({__name:"DataPlaneList",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{},gateways:{type:Boolean,default:!1}},emits:["load-data","change"],setup(n,{emit:i}){const a=n,{t:r,formatIsoDate:s}=ue(),p=ce()("use zones");function b(g){return g.map(u=>{var B,D,x,j,t,l;const T=u.mesh,o=u.name,w=((B=u.dataplane.networking.gateway)==null?void 0:B.type)||"STANDARD",O={name:w==="STANDARD"?"data-plane-detail-view":"gateway-detail-view",params:{mesh:T,dataPlane:o}},V=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],I=Ue(u.dataplane).filter(e=>V.includes(e.label)),N=(D=I.find(e=>e.label==="kuma.io/service"))==null?void 0:D.value,Q=(x=I.find(e=>e.label==="kuma.io/protocol"))==null?void 0:x.value,U=(j=I.find(e=>e.label==="kuma.io/zone"))==null?void 0:j.value;let M;N!==void 0&&(M={name:"service-detail-view",params:{mesh:T,service:N}});let P;U!==void 0&&(P={name:"zone-cp-detail-view",params:{zone:U}});const{status:F}=De(u.dataplane,u.dataplaneInsight),Z=((t=u.dataplaneInsight)==null?void 0:t.subscriptions)??[],H={totalUpdates:0,totalRejectedUpdates:0,dpVersion:null,envoyVersion:null,selectedTime:NaN,selectedUpdateTime:NaN,version:null},v=Z.reduce((e,y)=>{var J,X;if(y.connectTime){const Y=Date.parse(y.connectTime);(!e.selectedTime||Y>e.selectedTime)&&(e.selectedTime=Y)}const G=Date.parse(y.status.lastUpdateTime);return G&&(!e.selectedUpdateTime||G>e.selectedUpdateTime)&&(e.selectedUpdateTime=G),{totalUpdates:e.totalUpdates+parseInt(y.status.total.responsesSent??"0",10),totalRejectedUpdates:e.totalRejectedUpdates+parseInt(y.status.total.responsesRejected??"0",10),dpVersion:((J=y.version)==null?void 0:J.kumaDp.version)||e.dpVersion,envoyVersion:((X=y.version)==null?void 0:X.envoy.version)||e.envoyVersion,selectedTime:e.selectedTime,selectedUpdateTime:e.selectedUpdateTime,version:y.version||e.version}},H),L={name:o,detailViewRoute:O,type:w,zone:{title:U??r("common.collection.none"),route:P},service:{title:N??r("common.collection.none"),route:M},protocol:Q??r("common.collection.none"),status:F,totalUpdates:v.totalUpdates,totalRejectedUpdates:v.totalRejectedUpdates,envoyVersion:v.envoyVersion??r("common.collection.none"),warnings:[],lastUpdated:v.selectedUpdateTime?s(new Date(v.selectedUpdateTime).toUTCString()):r("common.collection.none"),lastConnected:v.selectedTime?s(new Date(v.selectedTime).toUTCString()):r("common.collection.none"),overview:u};if(v.version){const{kind:e}=Ne(v.version);e!==Le&&L.warnings.push(e)}return p&&v.dpVersion&&I.find(y=>y.label===Ie)&&typeof((l=v.version)==null?void 0:l.kumaDp.kumaCpCompatible)=="boolean"&&!v.version.kumaDp.kumaCpCompatible&&L.warnings.push(ze),L})}return(g,u)=>{const T=de("RouterLink");return f(),z(be,{"empty-state-message":c(r)("common.emptyState.message",{type:a.gateways?"Gateways":"Data Plane Proxies"}),"empty-state-cta-to":c(r)(`data-planes.href.docs.${a.gateways?"gateway":"data_plane_proxy"}`),"empty-state-cta-text":c(r)("common.documentation"),headers:[{label:"Name",key:"name"},...a.gateways?[{label:"Type",key:"type"}]:[],{label:"Service",key:"service"},...a.gateways?[]:[{label:"Protocol",key:"protocol"}],...c(p)?[{label:"Zone",key:"zone"}]:[],{label:"Last Updated",key:"lastUpdated"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":a.pageNumber,"page-size":a.pageSize,total:a.total,items:a.items?b(a.items):void 0,error:a.error,onChange:u[0]||(u[0]=o=>i("change",o))},{toolbar:h(()=>[ne(g.$slots,"toolbar",{},void 0,!0)]),name:h(({row:o})=>[C(T,{to:{name:a.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:o.name}},"data-testid":"detail-view-link"},{default:h(()=>[m(k(o.name),1)]),_:2},1032,["to"])]),service:h(({rowValue:o})=>[o.route?(f(),z(T,{key:0,to:o.route},{default:h(()=>[m(k(o.title),1)]),_:2},1032,["to"])):(f(),_(E,{key:1},[m(k(o.title),1)],64))]),zone:h(({rowValue:o})=>[o.route?(f(),z(T,{key:0,to:o.route},{default:h(()=>[m(k(o.title),1)]),_:2},1032,["to"])):(f(),_(E,{key:1},[m(k(o.title),1)],64))]),status:h(({rowValue:o})=>[o?(f(),z(pe,{key:0,status:o},null,8,["status"])):(f(),_(E,{key:1},[m(k(c(r)("common.collection.none")),1)],64))]),warnings:h(({rowValue:o})=>[o.length>0?(f(),z(c(me),{key:0,label:c(r)("data-planes.list.version_mismatch")},{default:h(()=>[C(fe,{class:"mr-1",size:c(R),"hide-title":""},null,8,["size"])]),_:1},8,["label"])):(f(),_(E,{key:1},[m(`
         
      `)],64))]),actions:h(({row:o})=>[C(c(ge),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:h(()=>[C(c(ve),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:h(()=>[C(c($),{color:c(ye),icon:"more",size:c(R)},null,8,["color","size"])]),_:1})]),items:h(()=>[C(c(he),{item:{to:{name:a.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:o.name}},label:c(r)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:3},8,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error"])}}});const nt=se(Ee,[["__scopeId","data-v-b3f65dc5"]]);function Re(n,i,a){return Math.max(i,Math.min(n,a))}const Me=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class Pe{constructor(i,a){K(this,"commands");K(this,"keyMap");K(this,"boundTriggerShortcuts");this.commands=a,this.keyMap=Object.fromEntries(Object.entries(i).map(([r,s])=>[r.toLowerCase(),s])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(i){Fe(i,this.keyMap,this.commands)}}function Fe(n,i,a){const r=Be(n.code),s=[n.ctrlKey?"ctrl":"",n.shiftKey?"shift":"",n.altKey?"alt":"",r].filter(b=>b!=="").join("+"),d=i[s];if(!d)return;const p=a[d];p.isAllowedContext&&!p.isAllowedContext(n)||(p.shouldPreventDefaultAction&&n.preventDefault(),!(p.isDisabled&&p.isDisabled())&&p.trigger(n))}function Be(n){return Me.includes(n)?"":n.replace(/^Key/,"").toLowerCase()}function je(n,i){const a=" "+n,r=a.matchAll(/ ([-\s\w]+):\s*/g),s=[];for(const d of Array.from(r)){if(d.index===void 0)continue;const p=Ke(d[1]);if(i.length>0&&!i.includes(p))throw new Error(`Unknown field “${p}”. Known fields: ${i.join(", ")}`);const b=d.index+d[0].length,g=a.substring(b);let u;if(/^\s*["']/.test(g)){const o=g.match(/['"](.*?)['"]/);if(o!==null)u=o[1];else throw new Error(`Quote mismatch for field “${p}”.`)}else{const o=g.indexOf(" "),w=o===-1?g.length:o;u=g.substring(0,w)}u!==""&&s.push([p,u])}return s}function Ke(n){return n.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(i,a)=>a===0?i:i.substring(1).toUpperCase())}let ae=0;const qe=(n="unique")=>(ae++,`${n}-${ae}`),ie=n=>(xe("data-v-e5b88bf8"),n=n(),Ae(),n),$e=ie(()=>S("span",{class:"visually-hidden"},"Focus filter",-1)),Oe=["for"],Ve=["id","placeholder"],Qe={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},Ze={class:"k-suggestion-list"},He={key:0,class:"k-filter-bar-error"},Ge={key:0},We=["title","data-filter-field"],Je={class:"visually-hidden"},Xe=ie(()=>S("span",{class:"visually-hidden"},"Clear query",-1)),Ye=oe({__name:"KFilterBar",props:{id:{type:String,required:!1,default:()=>qe("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(n,{emit:i}){const a=n,r=A(null),s=A(null),d=A(a.query),p=A([]),b=A(null),g=A(!1),u=A(-1),T=q(()=>Object.keys(a.fields)),o=q(()=>Object.entries(a.fields).slice(0,5).map(([t,l])=>({fieldName:t,...l}))),w=q(()=>T.value.length>0?`Filter by ${T.value.join(", ")}`:"Filter"),O=q(()=>a.placeholder??w.value);ee(()=>p.value,function(t,l){j(t,l)||(b.value=null,i("fields-change",{fields:t,query:d.value}))}),ee(()=>d.value,function(){d.value===""&&(b.value=null),g.value=!0});const V={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},I={submitQuery:{trigger:U,isAllowedContext(t){return s.value!==null&&t.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:M,isAllowedContext(t){return s.value!==null&&t.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:P,isAllowedContext(t){return s.value!==null&&t.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:D,isAllowedContext(t){return r.value!==null&&t.composedPath().includes(r.value)}}};function N(){const t=new Pe(V,I);Ce(function(){t.registerListener()}),we(function(){t.unRegisterListener()}),x(d.value)}N();function Q(t){const l=t.target;x(l.value)}function U(){if(s.value instanceof HTMLInputElement)if(u.value===-1)x(s.value.value),g.value=!1;else{const t=o.value[u.value].fieldName;t&&v(s.value,t)}}function M(){F(1)}function P(){F(-1)}function F(t){u.value=Re(u.value+t,-1,o.value.length-1)}function Z(){s.value instanceof HTMLInputElement&&s.value.focus()}function H(t){const e=t.currentTarget.getAttribute("data-filter-field");e&&s.value instanceof HTMLInputElement&&v(s.value,e)}function v(t,l){const e=d.value===""||d.value.endsWith(" ")?"":" ";d.value+=e+l+":",t.focus(),u.value=-1}function L(){d.value="",s.value instanceof HTMLInputElement&&(s.value.value="",s.value.focus(),x(""))}function B(t){t.relatedTarget===null&&D(),r.value instanceof HTMLElement&&t.relatedTarget instanceof Node&&!r.value.contains(t.relatedTarget)&&D()}function D(){g.value=!1}function x(t){b.value=null;try{const l=je(t,T.value);l.sort((e,y)=>e[0].localeCompare(y[0])),p.value=l}catch(l){if(l instanceof Error)b.value=l,g.value=!0;else throw l}}function j(t,l){return JSON.stringify(t)===JSON.stringify(l)}return(t,l)=>(f(),_("div",{ref_key:"filterBar",ref:r,class:"k-filter-bar","data-testid":"k-filter-bar"},[S("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:Z},[$e,m(),C(c($),{"aria-hidden":"true",class:"k-filter-icon",color:c(_e),"data-testid":"k-filter-bar-filter-icon","hide-title":"",icon:"filter",size:c(R)},null,8,["color","size"])]),m(),S("label",{for:`${a.id}-filter-bar-input`,class:"visually-hidden"},[ne(t.$slots,"default",{},()=>[m(k(w.value),1)],!0)],8,Oe),m(),ke(S("input",{id:`${a.id}-filter-bar-input`,ref_key:"filterInput",ref:s,"onUpdate:modelValue":l[0]||(l[0]=e=>d.value=e),class:"k-filter-bar-input",type:"text",placeholder:O.value,"data-testid":"k-filter-bar-filter-input",onFocus:l[1]||(l[1]=e=>g.value=!0),onBlur:B,onChange:Q},null,40,Ve),[[Te,d.value]]),m(),g.value?(f(),_("div",Qe,[S("div",Ze,[b.value!==null?(f(),_("p",He,k(b.value.message),1)):(f(),_("button",{key:1,class:te(["k-submit-query-button",{"k-submit-query-button-is-selected":u.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:U},`
          Submit `+k(d.value),3)),m(),(f(!0),_(E,null,Se(o.value,(e,y)=>(f(),_("div",{key:`${a.id}-${y}`,class:te(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":u.value===y}])},[S("b",null,k(e.fieldName),1),e.description!==""?(f(),_("span",Ge,": "+k(e.description),1)):W("",!0),m(),S("button",{class:"k-apply-suggestion-button",title:`Add ${e.fieldName}:`,type:"button","data-filter-field":e.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:H},[S("span",Je,"Add "+k(e.fieldName)+":",1),m(),C(c($),{"aria-hidden":"true",color:"currentColor","hide-title":"",icon:"chevronRight",size:c(R)},null,8,["size"])],8,We)],2))),128))])])):W("",!0),m(),d.value!==""?(f(),_("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:L},[Xe,m(),C(c($),{"aria-hidden":"true",color:"currentColor",icon:"clear","hide-title":"",size:c(R)},null,8,["size"])])):W("",!0)],512))}});const st=se(Ye,[["__scopeId","data-v-e5b88bf8"]]);export{nt as D,st as K};
