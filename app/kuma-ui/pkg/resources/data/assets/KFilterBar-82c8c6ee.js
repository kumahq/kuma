var le=Object.defineProperty;var re=(n,i,a)=>i in n?le(n,i,{enumerable:!0,configurable:!0,writable:!0,value:a}):n[i]=a;var q=(n,i,a)=>(re(n,typeof i!="symbol"?i+"":i,a),a);import{g as ue,S as ce,K as R,r as de,q as pe,f as oe,x as me}from"./RouteView.vue_vue_type_script_setup_true_lang-da83f5a8.js";import{d as ne,y as fe,r as ge,o as f,a as E,w as h,C as se,h as w,g as m,t as k,e as _,F as z,b as c,M as ve,R as ye,H as he,x as $,J as be,L as _e,l as x,c as K,Z as ee,i as S,ad as ke,ae as Te,n as te,v as Se,f as J,m as we,D as Ce,q as Ae,s as xe}from"./index-9a3d231d.js";import{_ as De}from"./WarningIcon.vue_vue_type_script_setup_true_lang-fe937ad6.js";import{d as Ne,a as Ue,c as Ie,C as Le,e as Ee}from"./dataplane-30467516.js";import{n as ze}from"./notEmpty-7f452b20.js";const Re=ne({__name:"DataPlaneList",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{},gateways:{type:Boolean,default:!1}},emits:["load-data","change"],setup(n,{emit:i}){const a=n,{t:r,formatIsoDate:s}=ue(),p=fe()("use zones");function b(g){return g.map(u=>{var B,U,A,j,t,l;const T=u.mesh,o=u.name,C=((B=u.dataplane.networking.gateway)==null?void 0:B.type)||"STANDARD",O={name:C==="STANDARD"?"data-plane-detail-view":"gateway-detail-view",params:{mesh:T,dataPlane:o}},V=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],D=Ne(u.dataplane).filter(e=>V.includes(e.label)),I=(U=D.find(e=>e.label==="kuma.io/service"))==null?void 0:U.value,Q=(A=D.find(e=>e.label==="kuma.io/protocol"))==null?void 0:A.value,N=(j=D.find(e=>e.label==="kuma.io/zone"))==null?void 0:j.value;let M;I!==void 0&&(M={name:"service-detail-view",params:{mesh:T,service:I}});let F;N!==void 0&&(F={name:"zone-cp-detail-view",params:{zone:N}});const{status:P}=Ue(u.dataplane,u.dataplaneInsight),H=((t=u.dataplaneInsight)==null?void 0:t.subscriptions)??[],Z={totalUpdates:0,totalRejectedUpdates:0,dpVersion:null,envoyVersion:null,selectedTime:NaN,selectedUpdateTime:NaN,version:null},v=H.reduce((e,y)=>{var W,X;if(y.connectTime){const Y=Date.parse(y.connectTime);(!e.selectedTime||Y>e.selectedTime)&&(e.selectedTime=Y)}const G=Date.parse(y.status.lastUpdateTime);return G&&(!e.selectedUpdateTime||G>e.selectedUpdateTime)&&(e.selectedUpdateTime=G),{totalUpdates:e.totalUpdates+parseInt(y.status.total.responsesSent??"0",10),totalRejectedUpdates:e.totalRejectedUpdates+parseInt(y.status.total.responsesRejected??"0",10),dpVersion:((W=y.version)==null?void 0:W.kumaDp.version)||e.dpVersion,envoyVersion:((X=y.version)==null?void 0:X.envoy.version)||e.envoyVersion,selectedTime:e.selectedTime,selectedUpdateTime:e.selectedUpdateTime,version:y.version||e.version}},Z),L={name:o,detailViewRoute:O,type:C,zone:{title:N??r("common.collection.none"),route:F},service:{title:I??r("common.collection.none"),route:M},protocol:Q??r("common.collection.none"),status:P,totalUpdates:v.totalUpdates,totalRejectedUpdates:v.totalRejectedUpdates,envoyVersion:v.envoyVersion??r("common.collection.none"),warnings:[],lastUpdated:v.selectedUpdateTime?s(new Date(v.selectedUpdateTime).toUTCString()):r("common.collection.none"),lastConnected:v.selectedTime?s(new Date(v.selectedTime).toUTCString()):r("common.collection.none"),overview:u};if(v.version){const{kind:e}=Ie(v.version);e!==Le&&L.warnings.push(e)}return p&&v.dpVersion&&D.find(y=>y.label===_e)&&typeof((l=v.version)==null?void 0:l.kumaDp.kumaCpCompatible)=="boolean"&&!v.version.kumaDp.kumaCpCompatible&&L.warnings.push(Ee),L})}return(g,u)=>{const T=ge("RouterLink");return f(),E(pe,{"empty-state-message":c(r)("common.emptyState.message",{type:a.gateways?"Gateways":"Data Plane Proxies"}),"empty-state-cta-to":c(r)(`data-planes.href.docs.${a.gateways?"gateway":"data_plane_proxy"}`),"empty-state-cta-text":c(r)("common.documentation"),headers:[{label:"Name",key:"name"},a.gateways?{label:"Type",key:"type"}:void 0,{label:"Service",key:"service"},a.gateways?void 0:{label:"Protocol",key:"protocol"},c(p)?{label:"Zone",key:"zone"}:void 0,{label:"Last Updated",key:"lastUpdated"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}].filter(c(ze)),"page-number":a.pageNumber,"page-size":a.pageSize,total:a.total,items:a.items?b(a.items):void 0,error:a.error,onChange:u[0]||(u[0]=o=>i("change",o))},{toolbar:h(()=>[se(g.$slots,"toolbar",{},void 0,!0)]),name:h(({row:o})=>[w(T,{to:{name:a.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:o.name}},"data-testid":"detail-view-link"},{default:h(()=>[m(k(o.name),1)]),_:2},1032,["to"])]),service:h(({rowValue:o})=>[o.route?(f(),E(T,{key:0,to:o.route},{default:h(()=>[m(k(o.title),1)]),_:2},1032,["to"])):(f(),_(z,{key:1},[m(k(o.title),1)],64))]),zone:h(({rowValue:o})=>[o.route?(f(),E(T,{key:0,to:o.route},{default:h(()=>[m(k(o.title),1)]),_:2},1032,["to"])):(f(),_(z,{key:1},[m(k(o.title),1)],64))]),status:h(({rowValue:o})=>[o?(f(),E(ce,{key:0,status:o},null,8,["status"])):(f(),_(z,{key:1},[m(k(c(r)("common.collection.none")),1)],64))]),warnings:h(({rowValue:o})=>[o.length>0?(f(),E(c(ve),{key:0,label:c(r)("data-planes.list.version_mismatch")},{default:h(()=>[w(De,{class:"mr-1",size:c(R),"hide-title":""},null,8,["size"])]),_:1},8,["label"])):(f(),_(z,{key:1},[m(`
         
      `)],64))]),actions:h(({row:o})=>[w(c(ye),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:h(()=>[w(c(he),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:h(()=>[w(c($),{color:c(de),icon:"more",size:c(R)},null,8,["color","size"])]),_:1})]),items:h(()=>[w(c(be),{item:{to:{name:a.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:o.name}},label:c(r)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:3},8,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error"])}}});const lt=oe(Re,[["__scopeId","data-v-67dbd621"]]);function Me(n,i,a){return Math.max(i,Math.min(n,a))}const Fe=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class Pe{constructor(i,a){q(this,"commands");q(this,"keyMap");q(this,"boundTriggerShortcuts");this.commands=a,this.keyMap=Object.fromEntries(Object.entries(i).map(([r,s])=>[r.toLowerCase(),s])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(i){Be(i,this.keyMap,this.commands)}}function Be(n,i,a){const r=je(n.code),s=[n.ctrlKey?"ctrl":"",n.shiftKey?"shift":"",n.altKey?"alt":"",r].filter(b=>b!=="").join("+"),d=i[s];if(!d)return;const p=a[d];p.isAllowedContext&&!p.isAllowedContext(n)||(p.shouldPreventDefaultAction&&n.preventDefault(),!(p.isDisabled&&p.isDisabled())&&p.trigger(n))}function je(n){return Fe.includes(n)?"":n.replace(/^Key/,"").toLowerCase()}function qe(n,i){const a=" "+n,r=a.matchAll(/ ([-\s\w]+):\s*/g),s=[];for(const d of Array.from(r)){if(d.index===void 0)continue;const p=Ke(d[1]);if(i.length>0&&!i.includes(p))throw new Error(`Unknown field “${p}”. Known fields: ${i.join(", ")}`);const b=d.index+d[0].length,g=a.substring(b);let u;if(/^\s*["']/.test(g)){const o=g.match(/['"](.*?)['"]/);if(o!==null)u=o[1];else throw new Error(`Quote mismatch for field “${p}”.`)}else{const o=g.indexOf(" "),C=o===-1?g.length:o;u=g.substring(0,C)}u!==""&&s.push([p,u])}return s}function Ke(n){return n.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(i,a)=>a===0?i:i.substring(1).toUpperCase())}let ae=0;const $e=(n="unique")=>(ae++,`${n}-${ae}`),ie=n=>(Ae("data-v-e5b88bf8"),n=n(),xe(),n),Oe=ie(()=>S("span",{class:"visually-hidden"},"Focus filter",-1)),Ve=["for"],Qe=["id","placeholder"],He={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},Ze={class:"k-suggestion-list"},Ge={key:0,class:"k-filter-bar-error"},Je={key:0},We=["title","data-filter-field"],Xe={class:"visually-hidden"},Ye=ie(()=>S("span",{class:"visually-hidden"},"Clear query",-1)),et=ne({__name:"KFilterBar",props:{id:{type:String,required:!1,default:()=>$e("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(n,{emit:i}){const a=n,r=x(null),s=x(null),d=x(a.query),p=x([]),b=x(null),g=x(!1),u=x(-1),T=K(()=>Object.keys(a.fields)),o=K(()=>Object.entries(a.fields).slice(0,5).map(([t,l])=>({fieldName:t,...l}))),C=K(()=>T.value.length>0?`Filter by ${T.value.join(", ")}`:"Filter"),O=K(()=>a.placeholder??C.value);ee(()=>p.value,function(t,l){j(t,l)||(b.value=null,i("fields-change",{fields:t,query:d.value}))}),ee(()=>d.value,function(){d.value===""&&(b.value=null),g.value=!0});const V={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},D={submitQuery:{trigger:N,isAllowedContext(t){return s.value!==null&&t.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:M,isAllowedContext(t){return s.value!==null&&t.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:F,isAllowedContext(t){return s.value!==null&&t.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:U,isAllowedContext(t){return r.value!==null&&t.composedPath().includes(r.value)}}};function I(){const t=new Pe(V,D);we(function(){t.registerListener()}),Ce(function(){t.unRegisterListener()}),A(d.value)}I();function Q(t){const l=t.target;A(l.value)}function N(){if(s.value instanceof HTMLInputElement)if(u.value===-1)A(s.value.value),g.value=!1;else{const t=o.value[u.value].fieldName;t&&v(s.value,t)}}function M(){P(1)}function F(){P(-1)}function P(t){u.value=Me(u.value+t,-1,o.value.length-1)}function H(){s.value instanceof HTMLInputElement&&s.value.focus()}function Z(t){const e=t.currentTarget.getAttribute("data-filter-field");e&&s.value instanceof HTMLInputElement&&v(s.value,e)}function v(t,l){const e=d.value===""||d.value.endsWith(" ")?"":" ";d.value+=e+l+":",t.focus(),u.value=-1}function L(){d.value="",s.value instanceof HTMLInputElement&&(s.value.value="",s.value.focus(),A(""))}function B(t){t.relatedTarget===null&&U(),r.value instanceof HTMLElement&&t.relatedTarget instanceof Node&&!r.value.contains(t.relatedTarget)&&U()}function U(){g.value=!1}function A(t){b.value=null;try{const l=qe(t,T.value);l.sort((e,y)=>e[0].localeCompare(y[0])),p.value=l}catch(l){if(l instanceof Error)b.value=l,g.value=!0;else throw l}}function j(t,l){return JSON.stringify(t)===JSON.stringify(l)}return(t,l)=>(f(),_("div",{ref_key:"filterBar",ref:r,class:"k-filter-bar","data-testid":"k-filter-bar"},[S("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:H},[Oe,m(),w(c($),{"aria-hidden":"true",class:"k-filter-icon",color:c(me),"data-testid":"k-filter-bar-filter-icon","hide-title":"",icon:"filter",size:c(R)},null,8,["color","size"])]),m(),S("label",{for:`${a.id}-filter-bar-input`,class:"visually-hidden"},[se(t.$slots,"default",{},()=>[m(k(C.value),1)],!0)],8,Ve),m(),ke(S("input",{id:`${a.id}-filter-bar-input`,ref_key:"filterInput",ref:s,"onUpdate:modelValue":l[0]||(l[0]=e=>d.value=e),class:"k-filter-bar-input",type:"text",placeholder:O.value,"data-testid":"k-filter-bar-filter-input",onFocus:l[1]||(l[1]=e=>g.value=!0),onBlur:B,onChange:Q},null,40,Qe),[[Te,d.value]]),m(),g.value?(f(),_("div",He,[S("div",Ze,[b.value!==null?(f(),_("p",Ge,k(b.value.message),1)):(f(),_("button",{key:1,class:te(["k-submit-query-button",{"k-submit-query-button-is-selected":u.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:N},`
          Submit `+k(d.value),3)),m(),(f(!0),_(z,null,Se(o.value,(e,y)=>(f(),_("div",{key:`${a.id}-${y}`,class:te(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":u.value===y}])},[S("b",null,k(e.fieldName),1),e.description!==""?(f(),_("span",Je,": "+k(e.description),1)):J("",!0),m(),S("button",{class:"k-apply-suggestion-button",title:`Add ${e.fieldName}:`,type:"button","data-filter-field":e.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:Z},[S("span",Xe,"Add "+k(e.fieldName)+":",1),m(),w(c($),{"aria-hidden":"true",color:"currentColor","hide-title":"",icon:"chevronRight",size:c(R)},null,8,["size"])],8,We)],2))),128))])])):J("",!0),m(),d.value!==""?(f(),_("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:L},[Ye,m(),w(c($),{"aria-hidden":"true",color:"currentColor",icon:"clear","hide-title":"",size:c(R)},null,8,["size"])])):J("",!0)],512))}});const rt=oe(et,[["__scopeId","data-v-e5b88bf8"]]);export{lt as D,rt as K};
